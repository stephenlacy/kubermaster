package manager

import (
	// "flag"
	"fmt"
	"time"

	"github.com/julienschmidt/httprouter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	// "k8s.io/client-go/tools/clientcmd"
	"net/http"
	// "path/filepath"
)

// DefaultMemory is the maximum memory assigned to a Pod
var DefaultMemory = "650Mi"

// DefaultCPURequest is the cpu amount requested
var DefaultCPURequest = "0.22"

// DefaultCPULimit is the maximum cpu amount
var DefaultCPULimit = "0.44"

// RootToken is the token used for the cluster's authentication
var RootToken string

// PostRequest is the request sent to the manager
type PostRequest struct {
	Token      string            `json:"token"`
	Command    string            `json:"command"`
	Image      string            `json:"image"`
	Auth       string            `json:"auth"`
	Labels     map[string]string `json:"labels`
	Name       string            `json:"name"`
	Id         string            `json:"id"`
	Memory     string            `json:"memory"` // 1G
	CPULimit   string            `json:"cpuLimit"`
	CPURequest string            `json:"cpuRequest"`
	JobID      string            `json:"jobId"`
	ImporterID string            `json:"importerId"`
	SourceID   string            `json:"sourceId"`
	PreStop    string            `json:"preStop"`
}

// PostSuccessResponse is the JSON success response payload
type PostSuccessResponse struct {
	Success    bool   `json:"success"`
	Id         string `json:"id"`
	ImporterID string `json:"importerId"`
	Status     string `json:"status"`
}

// JobStatus is the JSON status sent back to the api
type JobStatus struct {
	Status string `json:"status"`
}

// PostErrorResponse is the JSON error response payload
type PostErrorResponse struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Error   string `json:"error"`
	Auth    bool   `json:"auth"`
	Id      string `json:"id"`
}

// Init starts the manager
func Init(token string, memory string) http.Handler {
	router := httprouter.New()
	RootToken = token

	var err error

	if memory != "" {
		DefaultMemory = memory
	}

	// inside the cluster:
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// var kubeconfig *string
	// if home := os.Getenv("HOME"); home != "" {
	// 	kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	// } else {
	// 	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	// }
	// flag.Parse()
	//
	// // use the current context in kubeconfig
	// config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	// if err != nil {
	// 	panic(err.Error())
	// }

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	listOptions := metav1.ListOptions{
		LabelSelector: "type=importer",
	}

	tasks, err := clientset.Core().Pods(metav1.NamespaceDefault).List(listOptions)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d tasks in the cluster\n", len(tasks.Items))

	// Every 5 minutes do a cleanup of all old Pods. This increases performance on all 1.11* kubernetes versions
	ticker := time.NewTicker(30 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				PurgeSelector(*clientset, "status.phase=Failed")
				PurgeSelector(*clientset, "status.phase=Succeeded")
			}
		}
	}()

	router.POST("/run", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		response, err := HandlePostAuth(w, r)
		r.Close = true
		r.Header.Set("Content-Type", "application/json")
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		Run(w, r, p, response, *clientset)
	})

	router.POST("/stop", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		response, err := HandlePostAuth(w, r)
		r.Close = true
		r.Header.Set("Content-Type", "application/json")
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		Stop(w, r, p, response, *clientset)
	})

	router.POST("/purge", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		response, err := HandlePostAuth(w, r)
		r.Close = true
		r.Header.Set("Content-Type", "application/json")
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		Purge(w, r, p, response, *clientset)
	})

	router.GET("/status", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		r.Header.Set("Content-Type", "application/json")
		w.Header().Set("Content-Type", "application/json")
		response, err := HandleGetAuth(w, r)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		Status(w, r, p, response, *clientset)
	})

	// router.GET("/status/:id", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// 	response, err := HandleGetAuth(w, r)
	// 	if err != nil {
	// 		fmt.Fprint(w, err)
	// 		return
	// 	}
	// 	Status(w, r, p, response, *clientset)
	// })

	return router
}
