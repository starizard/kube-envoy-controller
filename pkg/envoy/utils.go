package envoy

import (
	"encoding/json"
	"log"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	v1 "github.com/starizard/kube-envoy-controller/pkg/api/example.com/v1"
	client "github.com/starizard/kube-envoy-controller/pkg/client/clientset/versioned"
)

var apiType = "GRPC"

//Deployment returns a spec for an envoy deployment
func Deployment(envoy *v1.Envoy) *appsv1.Deployment {
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
							Name:    "envoy",
							Image:   "envoyproxy/envoy:v1.10.0",
							Command: []string{"envoy"},
							Args:    []string{"-c", "/etc/envoy.yaml"},
							VolumeMounts: []apiv1.VolumeMount{
								apiv1.VolumeMount{
									Name:      "envoy-yaml",
									MountPath: "/etc/envoy.yaml",
									SubPath:   "envoy.yaml",
								},
							},
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 8080,
								},
							},
						},
					},
					Volumes: []apiv1.Volume{
						apiv1.Volume{
							Name: "envoy-yaml",
							VolumeSource: apiv1.VolumeSource{
								ConfigMap: &apiv1.ConfigMapVolumeSource{
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: envoy.Spec.ConfigMapName,
									},
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

//Service returns a spec for an envoy service
func Service(envoy *v1.Envoy) *apiv1.Service {
	service := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: envoy.Spec.Name,
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				apiv1.ServicePort{
					Name:       "http",
					Protocol:   "TCP",
					Port:       80,
					TargetPort: intstr.IntOrString{IntVal: 8080},
				},
			},
			Selector: map[string]string{
				"app": "envoy",
			},
		},
	}
	return service
}

func addAdminConfig() Admin {
	return Admin{
		AccessLogPath: "/dev/stderr",
		Address: Address{
			SocketAddress: SocketAddress{
				Address:   "127.0.0.1",
				PortValue: 15000,
			},
		},
	}
}

func addLDSConfig(envoy *v1.Envoy) LdsConfig {
	clusterName := envoy.Spec.XDS.Name
	return LdsConfig{
		APIConfigSource: APIConfigSource{
			APIType: apiType,
			GrpcServices: GrpcServices{
				Grpc: Grpc{
					ClusterName: clusterName,
				},
			},
		},
	}
}

func addCDSConfig(envoy *v1.Envoy) CdsConfig {
	clusterName := envoy.Spec.XDS.Name
	return CdsConfig{
		APIConfigSource: APIConfigSource{
			APIType: apiType,
			GrpcServices: GrpcServices{
				Grpc: Grpc{
					ClusterName: clusterName,
				},
			},
		},
	}
}

func addADSConfig(envoy *v1.Envoy) AdsConfig {
	clusterName := envoy.Spec.XDS.Name
	return AdsConfig{
		APIType: apiType,
		GrpcServices: GrpcServices{
			Grpc: Grpc{
				ClusterName: clusterName,
			},
		},
	}
}

func addDynamicResources(envoy *v1.Envoy) DynamicResources {
	return DynamicResources{
		AdsConfig: addADSConfig(envoy),
		CdsConfig: addCDSConfig(envoy),
		LdsConfig: addLDSConfig(envoy),
	}
}

func addStaticResources(envoy *v1.Envoy) StaticResources {
	clusterName := envoy.Spec.XDS.Name
	xdsHostname := envoy.Spec.XDS.Host
	xdsPort := envoy.Spec.XDS.Port
	return StaticResources{
		Clusters: []Clusters{{
			Name:           clusterName,
			ConnectTimeout: "5s",
			Type:           "STRICT_DNS",
			Hosts: []Hosts{{
				SocketAddress: SocketAddress{
					Address:   xdsHostname,
					PortValue: xdsPort,
				},
			},
			},
		},
		},
	}
}

func makeEnvoyConfig(envoy *v1.Envoy) *Bootstrap {
	envoyconfig := &Bootstrap{
		Node: Node{
			Cluster: "service_1",
			ID:      "test-id",
		},

		StaticResources:  addStaticResources(envoy),
		DynamicResources: addDynamicResources(envoy),
		Admin:            addAdminConfig(),
	}
	return envoyconfig
}

//ConfigMap returns a spec for an envoy bootstrap config
func ConfigMap(envoy *v1.Envoy) *apiv1.ConfigMap {
	var cfgData string
	conf := makeEnvoyConfig(envoy)
	jsonString, err := json.Marshal(conf)
	if err != nil {
		log.Println(err)
	}

	cfgData = string(jsonString)
	log.Println(cfgData)

	data := map[string]string{
		"envoy.yaml": cfgData,
	}
	cfgMap := &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: envoy.Spec.ConfigMapName,
		},
		Data: data,
	}
	return cfgMap
}

//UpdateStatus updates the status of an envoy resource
func UpdateStatus(clientset client.Interface, envoy *v1.Envoy, namespace string, deployment *appsv1.Deployment) error {
	updatedObj := envoy.DeepCopy()
	updatedObj.Status.AvailableReplicas = deployment.Status.AvailableReplicas
	//TODO: use .UpdateStatus()? Might need a subresource
	_, err := clientset.ExampleV1().Envoys(namespace).Update(updatedObj)
	return err
}
