package agent

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	wfclientset "github.com/argoproj/argo/pkg/client/clientset/versioned"
	"github.com/argoproj/argo/workflow/util"
)

var kubeconfigPtr = flag.String("kubeconfig", filepath.Join(os.Getenv("HOME"), ".kube", "config"), "(optional) absolute path to the kubeconfig file")

// ArgoAPI represents the Argo API endpoint within the Kubernetes API
type ArgoAPI struct {
	client *wfclientset.Clientset
}

// NewArgoAPI creates a new ArgoAPI
func NewArgoAPI() ArgoAPI {
	// creates the in-cluster config
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfigPtr)

	if err != nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		log.Print("Running in-cluster mode")
	} else {
		log.Printf("Using kubernetes configuration from %s", *kubeconfigPtr)
	}
	wfClient := wfclientset.NewForConfigOrDie(config)

	return ArgoAPI{client: wfClient}

}

/*
ResumeWorkflow resumes workflows in namespace
*/
func (a *ArgoAPI) ResumeWorkflow(workflow string, namespace string) error {

	wfClient := a.client.ArgoprojV1alpha1().Workflows(namespace)
	return util.ResumeWorkflow(wfClient, workflow)
}
