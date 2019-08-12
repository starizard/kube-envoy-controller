package main

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"

	v1 "github.com/starizard/kube-envoy-controller/pkg/api/example.com/v1"
	client "github.com/starizard/kube-envoy-controller/pkg/client/clientset/versioned"
	factory "github.com/starizard/kube-envoy-controller/pkg/client/informers/externalversions"
	envoyutils "github.com/starizard/kube-envoy-controller/pkg/envoy"
)

var (
	queue         = workqueue.NewRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(time.Second*5, time.Minute))
	clientset     client.Interface
	kubeclientset kubernetes.Interface
	stopCh        = make(chan struct{})
	sharedFactory factory.SharedInformerFactory
)

func getConfig() *rest.Config {
	kubeconfig := ""
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
	return config
}

func createClientSet() *client.Clientset {
	return client.NewForConfigOrDie(getConfig())
}

func createKubeClientSet() *kubernetes.Clientset {
	return kubernetes.NewForConfigOrDie(getConfig())
}

func main() {
	clientset = createClientSet()
	kubeclientset = createKubeClientSet()
	// Create A shared informer factory, then use that factory to create a informer for your custom type
	sharedFactory = factory.NewSharedInformerFactory(clientset, time.Second*30)
	informer := sharedFactory.Example().V1().Envoys().Informer()

	// Add informer event handlers to respond to changes in the resource, we can enqueue the new changes to the workqueue
	informer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				fmt.Println("\nEnvoy Added\n")
				enqueue(obj)

			},
			UpdateFunc: func(old interface{}, cur interface{}) {
				if !reflect.DeepEqual(old, cur) {
					fmt.Println("\nEnvoy updated\n")
					enqueue(cur)

				}
			},
			DeleteFunc: func(obj interface{}) {
				fmt.Println("\nEnvoy deleted\n")
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

	//retrieve the object
	obj, err := sharedFactory.Example().V1().Envoys().Lister().Envoys(namespace).Get(name)
	if err != nil {
		fmt.Printf("\nError getting object %s %s from api %s", namespace, name, err)
	}

	//Reconcile expected state with current state
	if err := reconcile(obj, namespace, name); err != nil {
		fmt.Printf("\nError reconciling object %v", err)
		return
	}
}

func reconcile(envoy *v1.Envoy, namespace string, name string) error {
	fmt.Printf("\n Processing Envoy %s, %d, %s", envoy.Spec.Name, *envoy.Spec.Replicas, envoy.Spec.ConfigMapName)
	deploymentName := envoy.Spec.Name
	if deploymentName == "" {
		runtime.HandleError(fmt.Errorf("%s: deployment name must be specified", name))
		return nil
	}
	deploymentsClient := kubeclientset.AppsV1().Deployments(namespace)
	deployment, err := deploymentsClient.Get(envoy.Spec.Name, metav1.GetOptions{})

	if errors.IsNotFound(err) {
		fmt.Printf("Deployment not found %v", err)
		newDeploymentSpec := envoyutils.Deployment(envoy)
		deployment, _ = deploymentsClient.Create(newDeploymentSpec)

		// TODO: update envoy status
	}
	if err == nil {
		if envoy.Spec.Replicas != nil && *envoy.Spec.Replicas != *deployment.Spec.Replicas {
			_, _ = deploymentsClient.Update(envoyutils.Deployment(envoy))
			fmt.Printf("Updating deployments")
		}
	}
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
