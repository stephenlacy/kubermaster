package manager

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"net/http"
)

// Status returns the status of all current importers in the cluster
func Status(w http.ResponseWriter, r *http.Request, p httprouter.Params, response PostRequest, clientset kubernetes.Clientset) {
	podList := metav1.ListOptions{
	// LabelSelector: "importer",
	}

	results := []PostSuccessResponse{}

	tasks, err := clientset.Core().Pods(metav1.NamespaceDefault).List(podList)
	if err != nil {
		panic(err)
	}
	for _, task := range tasks.Items {
		success := false
		phase := fmt.Sprintf("%v", task.Status.Phase)
		if phase == "Running" {
			success = true
		}
		if phase == "Succeeded" {
			success = true
		}
		formatted := PostSuccessResponse{
			Id:      task.GetName(),
			Status:  phase,
			Success: success,
		}
		results = append(results, formatted)
		fmt.Printf("wat: %v", results)
	}
	_ = json.NewEncoder(w).Encode(results)
	return
}
