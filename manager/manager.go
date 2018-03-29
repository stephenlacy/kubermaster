package manager

import (
	// "flag"
	"fmt"
	"github.com/julienschmidt/httprouter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	// "k8s.io/client-go/tools/clientcmd"
	"net/http"
	// "os"
	// "path/filepath"
)

var root_token string = ""
var DEFAULT_MEMORY string = "250M"

type PostRequest struct {
	Token   string            `json:"token"`
	Command string            `json:"command"`
	Image   string            `json:"image"`
	Auth    string            `json:"auth"`
	Labels  map[string]string `json:"labels`
	Name    string            `json:"name"`
	Id      string            `json:"id"`
	Memory  string            `json:"memory"` // 250M
	JobID   string            `json:"jobId"`
	PreStop string            `json:"preStop"`
}

type PostSuccessResponse struct {
	Success bool   `json:"success"`
	Id      string `json:"id"`
	Status  string `json:"status"`
}

type PostErrorResponse struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Error   string `json:"error"`
	Auth    bool   `json:"auth"`
	Id      string `json:"id"`
}

func Init(token string, memory string) http.Handler {
	router := httprouter.New()
	root_token = token

	if memory != "" {
		DEFAULT_MEMORY = memory
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

	jobs, err := clientset.BatchV1().Jobs("").List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d jobs in the cluster\n", len(jobs.Items))

	router.POST("/run", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		response, err := HandlePostAuth(w, r)
		r.Close = true
		r.Header.Set("Content-Type", "application/json")
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
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		Stop(w, r, p, response, *clientset)
	})

	router.GET("/status", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
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
