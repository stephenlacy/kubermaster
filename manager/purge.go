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
	fmt.Printf("purging dead tasks %s\n", selector)
	propagationPolicy := metav1.DeletePropagationForeground

	deleteOptions := &metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	}
	runningListOptions := metav1.ListOptions{
		LabelSelector: "type=importer",
		FieldSelector: selector,
	}

	// Delete just the dead importers
	err := clientset.Core().Pods(metav1.NamespaceDefault).DeleteCollection(deleteOptions, runningListOptions)
	if err != nil {
		fmt.Printf("%e", err)
	}
	return
}
