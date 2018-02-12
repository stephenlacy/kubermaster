package manager

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	// "k8s.io/client-go/rest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Stop (delete) a job
func Stop(w http.ResponseWriter, r *http.Request, p httprouter.Params, response PostRequest, clientset kubernetes.Clientset) {
	if response.Id == "" {
		payload := PostErrorResponse{Success: false, Error: "container id missing"}
		_ = json.NewEncoder(w).Encode(payload)
		return
	}

	propagationPolicy := metav1.DeletePropagationBackground

	err := clientset.BatchV1().Jobs(metav1.NamespaceDefault).Delete(response.Id, &metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
	fmt.Printf("deleting job: %v", response.Id)

	if err != nil {
		fmt.Printf("Error deleting job: %v, error: %v", response.Id, err)
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

	payload := &PostSuccessResponse{Success: true, Id: response.Id}
	_ = json.NewEncoder(w).Encode(payload)
	return
}
