package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"strings"
)

// Purge (delete) a task
func Purge(w http.ResponseWriter, r *http.Request, p httprouter.Params, response PostRequest, clientset kubernetes.Clientset) {
	fmt.Println("purging tasks")
	propagationPolicy := metav1.DeletePropagationBackground

	deleteOptions := &metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	}
	runningListOptions := metav1.ListOptions{
		LabelSelector: "type=importer",
		FieldSelector: "status.phase==Running",
	}

	// Delete just the running importers
	err := clientset.Core().Pods(metav1.NamespaceDefault).DeleteCollection(deleteOptions, runningListOptions)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	payload := &PostSuccessResponse{Success: true}
	_ = json.NewEncoder(w).Encode(payload)
	return
}

// PurgeSelector purges tasks by selector
func PurgeSelector(clientset kubernetes.Clientset, selector string) {
	propagationPolicy := metav1.DeletePropagationForeground

	deleteOptions := &metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	}
	listOptions := metav1.ListOptions{
		LabelSelector: "type=importer",
		FieldSelector: selector,
	}

	tasks, err := clientset.Core().Pods(metav1.NamespaceDefault).List(listOptions)
	if err != nil {
		fmt.Printf("%e", err)
	}

	for _, task := range tasks.Items {
		secretKey := ""
		statusURL := ""
		for _, v := range task.Spec.Containers[0].Args {
			if strings.Contains(v, "--secret-key") {
				secretKey = strings.Replace(v, "--secret-key=", "", 1)
			}
			if strings.Contains(v, "--status-endpoint") {
				statusURL = strings.Replace(v, "--status-endpoint=", "", 1)
			}
		}
		if secretKey == "" || statusURL == "" {
			break
		}
		// Generate the full url
		url := fmt.Sprintf(
			"%s?secretKey=%s",
			statusURL,
			secretKey,
		)
		// Send request back to api te alert not running status

		status := "failed"
		if task.Status.Phase == "Succeeded" {
			status = "completed"
		}

		stat := JobStatus{
			Status: status,
		}

		makeRequest(url, stat)
		err = clientset.Core().Pods(metav1.NamespaceDefault).Delete(task.Name, deleteOptions)
		if err != nil {
			fmt.Printf("%e", err)
		}
	}
	return
}

func makeRequest(url string, stat JobStatus) {
	data, err := json.Marshal(&stat)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	txn := NewRelicClient.StartTransaction("importer:api:update", nil, req)
	defer txn.End()
	if err != nil {
		fmt.Printf("error creating request %s\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error connecting to api\n")
		return
	}
	defer resp.Body.Close()
}
