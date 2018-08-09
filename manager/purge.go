package manager

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"net/http"
)

// Purge (delete) a task
func Purge(w http.ResponseWriter, r *http.Request, p httprouter.Params, response PostRequest, clientset kubernetes.Clientset) {
	fmt.Println("purging tasks")
	propagationPolicy := metav1.DeletePropagationBackground

	deleteOptions := &metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	}
	runningListOptions := metav1.ListOptions{
		FieldSelector: "status.phase==Running",
	}
	// finalListOptions := metav1.ListOptions{
	// 	LabelSelector: "type=importer",
	// }
	// Delete just the running importers first
	err := clientset.Core().Pods(metav1.NamespaceDefault).DeleteCollection(deleteOptions, runningListOptions)

	if err != nil {
		fmt.Printf("Error purging tasks, error: %v", err)
		w.WriteHeader(400)
		payload := &PostErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    400,
		}
		_ = json.NewEncoder(w).Encode(payload)
		return
	}

	w.WriteHeader(200)
	payload := &PostSuccessResponse{Success: true}
	_ = json.NewEncoder(w).Encode(payload)
	return
}
