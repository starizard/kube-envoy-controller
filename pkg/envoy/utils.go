package envoy

import (
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/starizard/kube-envoy-controller/pkg/api/example.com/v1"
	client "github.com/starizard/kube-envoy-controller/pkg/client/clientset/versioned"
)

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

//UpdateStatus updates the status of an envoy resource
func UpdateStatus(clientset client.Interface, envoy *v1.Envoy, namespace string, deployment *appsv1.Deployment) error {
	updatedObj := envoy.DeepCopy()
	updatedObj.Status.AvailableReplicas = deployment.Status.AvailableReplicas
	//TODO: use .UpdateStatus()? Might need a subresource
	_, err := clientset.ExampleV1().Envoys(namespace).Update(updatedObj)
	return err
}
