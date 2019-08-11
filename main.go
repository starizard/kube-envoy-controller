package main

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"time"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
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
	// Create A shared informer factory, then use that factory to create a informer for your custom type
	sharedFactory = factory.NewSharedInformerFactory(cl, time.Second*30)
	informer := sharedFactory.Example().V1().Envoys().Informer()

	// Add informer event handlers to respond to changes in the resource, we can enqueue the new changes to the workqueue
	informer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				fmt.Println("Envoy Added")
				enqueue(obj)

			},
			UpdateFunc: func(old interface{}, cur interface{}) {
				if !reflect.DeepEqual(old, cur) {
					fmt.Println("Envoy updated")
					enqueue(cur)

				}
			},
			DeleteFunc: func(obj interface{}) {
				fmt.Println("Envoy deleted")
				enqueue(obj)
			},
		},
	)

	// this starts all registered informers
	sharedFactory.Start(stopCh)
	fmt.Println("Informer Started..")

	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		fmt.Println(("Error waiting for informer cache to sync"))
	}

}

func enqueue(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		fmt.Printf("Error obtaining key %v", err)
		return
	}

	queue.Add(key)
}
