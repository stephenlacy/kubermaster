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
		LabelSelector: "type=importer",
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
			Id:         task.GetName(),
			ImporterID: task.Labels["importerId"],
			Status:     phase,
			Success:    success,
		}
		exists := false
		for _, existing := range results {
			if existing.ImporterID == formatted.ImporterID && formatted.ImporterID != "" && existing.Status == formatted.Status {
				exists = true
			}
		}
		if !exists {
			results = append(results, formatted)
		}
	}
	_ = json.NewEncoder(w).Encode(results)
	return
}
