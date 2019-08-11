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

	v1 "github.com/starizard/kube-envoy-controller/pkg/api/example.com/v1"
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

	// Start controller loop
	work()
}

func work() {
	for {
		key, shutdown := queue.Get()

		if shutdown {
			stopCh <- struct{}{}
			return
		}
		var strKey string
		var ok bool
		if strKey, ok = key.(string); !ok {
			fmt.Printf("\n Invalid key format %v", key)
			return
		}
		processItem(strKey)
	}
}

func processItem(key string) {
	defer queue.Done(key)
	fmt.Printf("\nProcessItem %v", key)
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		fmt.Printf("\nError splitting key into parts %v", err)
		return
	}
	fmt.Printf("\nProcessing key %s %s", namespace, name)

	//retrieve the object
	obj, err := sharedFactory.Example().V1().Envoys().Lister().Envoys(namespace).Get(name)
	if err != nil {
		fmt.Printf("\nError getting object %s %s from api %s", namespace, name, err)
	}

	//Reconcile expected state with current state
	if err := reconcile(obj); err != nil {
		fmt.Printf("\nError reconciling object %v", err)

		return
	}
}

func reconcile(envoy *v1.Envoy) error {
	fmt.Printf("\n Processing Envoy %s, %d, %s", envoy.Spec.Name, *envoy.Spec.Replicas, envoy.Spec.ConfigMapName)

	return nil
}

func enqueue(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		fmt.Printf("Error obtaining key %v", err)
		return
	}

	queue.Add(key)
}
