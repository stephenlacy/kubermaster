package manager

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	api "k8s.io/api/core/v1"
	"net/http"
	"strings"
	// "k8s.io/client-go/rest"
	"fmt"
	"github.com/satori/go.uuid"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Run a pod
func Run(w http.ResponseWriter, r *http.Request, p httprouter.Params, response PostRequest, clientset kubernetes.Clientset) {
	if response.Command == "" {
		payload := PostErrorResponse{Success: false, Error: "command missing or invalid"}
		_ = json.NewEncoder(w).Encode(payload)
		return
	}
	if response.Image == "" {
		payload := PostErrorResponse{Success: false, Error: "image missing or invalid"}
		_ = json.NewEncoder(w).Encode(payload)
		return
	}

	id := ""
	if response.Name == "" {
		uid := uuid.NewV4()
		id = uid.String()
	}

	response.Name = fmt.Sprintf("job-%v-%v-%v", response.Name, id, api.NamespaceDefault)

	if response.Memory == "" {
		response.Memory = DEFAULT_MEMORY
	}

	args := strings.Split(response.Command, " ")

	imagePullSecrets := []api.LocalObjectReference{}
	imagePullSecrets = append(imagePullSecrets, api.LocalObjectReference{Name: "regsecret"})

	env := []api.EnvVar{}
	env = append(env, api.EnvVar{
		Name:  "JOB_ID",
		Value: response.JobID,
	})

	lifecycle := &api.Lifecycle{}

	if response.PreStop != "" {
		preStop := strings.Split(response.PreStop, " ")

		lifecycle = &api.Lifecycle{
			PreStop: &api.Handler{
				Exec: &api.ExecAction{
					Command: preStop,
				},
			},
		}
	}

	podSpec := api.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   response.Name,
			Labels: map[string]string{"type": "importer"},
		},
		Spec: api.PodSpec{
			RestartPolicy:    api.RestartPolicyNever,
			ImagePullSecrets: imagePullSecrets,
			Containers: []api.Container{
				{
					Name:            response.Name,
					Image:           response.Image,
					ImagePullPolicy: "Always",
					Command:         args[:0],
					Env:             env,
					Args:            args,
					Lifecycle:       lifecycle,
					Resources: api.ResourceRequirements{
						Limits: api.ResourceList{
							api.ResourceName(api.ResourceMemory): resource.MustParse(response.Memory),
						},
					},
				},
			},
		},
	}

	job, err := clientset.Core().Pods(metav1.NamespaceDefault).Create(&podSpec)
	fmt.Printf("Creating job: %v", job.GetName())

	if err != nil {
		fmt.Printf("Error creating job: %v, error: %v", job.GetName(), err)
		w.WriteHeader(400)
		payload := PostErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    400,
			Id:      response.Id,
		}
		_ = json.NewEncoder(w).Encode(payload)
		return
	}

	payload := &PostSuccessResponse{Success: true, Id: job.GetName()}
	_ = json.NewEncoder(w).Encode(payload)
	return
}
