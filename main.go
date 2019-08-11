package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"

	client "github.com/starizard/kube-envoy-controller/pkg/client/clientset/versioned"
	factory "github.com/starizard/kube-envoy-controller/pkg/client/informers/externalversions"
)

var (
	queue         = workqueue.NewRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(time.Second*5, time.Minute))
	cl            client.Interface
	stopCh        = make(chan struct{})
	sharedFactory factory.SharedInformerFactory
)

func createKubeClientSet() *client.Clientset {
	kubeconfig := ""
	flag.StringVar(&kubeconfig, "kubeconfig", kubeconfig, "kubeconfig file")
	if kubeconfig == "" {
		kubeconfig = os.Getenv("KUBECONFIG")
	}

	var (
		config *rest.Config
		err    error
	)

	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating client: %v", err)
		os.Exit(1)
	}

	return client.NewForConfigOrDie(config)
}

func main() {
	cl = createKubeClientSet()
}
