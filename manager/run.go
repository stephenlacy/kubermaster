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
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

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
		uid := uuid.NewV1()
		id = uid.String()
	}

	response.Name = fmt.Sprintf("job-%v-%v-%v", response.Name, id, api.NamespaceDefault)

	if response.Memory == "" {
		response.Memory = DEFAULT_MEMORY
	}

	args := strings.Split(response.Command, " ")

	imagePullSecrets := []api.LocalObjectReference{}
	imagePullSecrets = append(imagePullSecrets, api.LocalObjectReference{Name: "regsecret"})

	backOffLimit := int32(1)
	parallelisim := int32(1)
	completions := int32(1)

	jobTemplate := api.PodTemplateSpec{
		Spec: api.PodSpec{
			RestartPolicy:    api.RestartPolicyOnFailure,
			ImagePullSecrets: imagePullSecrets,
			Containers: []api.Container{
				{
					Name:            response.Name,
					Image:           response.Image,
					ImagePullPolicy: "IfNotPresent",
					Command:         args[:0],
					Args:            args,
					Resources: api.ResourceRequirements{
						Limits: api.ResourceList{
							api.ResourceName(api.ResourceMemory): resource.MustParse(response.Memory),
						},
					},
				},
			},
		},
	}
	jobopts := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      response.Name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: batchv1.JobSpec{
			Template:     jobTemplate,
			BackoffLimit: &backOffLimit,
			Parallelism:  &parallelisim,
			Completions:  &completions,
		},
	}
	job, err := clientset.BatchV1().Jobs(metav1.NamespaceDefault).Create(jobopts)
	if err != nil {
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
