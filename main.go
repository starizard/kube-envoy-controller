package main

import (
	"fmt"
	"os"
	"reflect"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
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
	_, err := deploymentsClient.Get(name, metav1.GetOptions{})

	if errors.IsNotFound(err) {
		deployment, _ := deploymentsClient.Create(newDeployment(envoy))
		if envoy.Spec.Replicas != nil && *envoy.Spec.Replicas != *deployment.Spec.Replicas {
			deployment, _ = deploymentsClient.Update(newDeployment(envoy))
		}
		// TODO: update envoy status
	}
	return nil
}

//TODO: move to other file
func newDeployment(envoy *v1.Envoy) *appsv1.Deployment {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: envoy.Spec.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: envoy.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "envoy",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "envoy",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "envoy",
							Image: "envoyproxy/envoy:v1.10.0",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	return deployment
}

func enqueue(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		fmt.Printf("Error obtaining key %v", err)
		return
	}

	queue.Add(key)
}
