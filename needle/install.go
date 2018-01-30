package needle

import (
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"

	appsv1beta1 "k8s.io/api/apps/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func IntOrStringPtr(i intstr.IntOrString) *intstr.IntOrString { return &i }

func Install(cc *kubernetes.Clientset) error {
	var replicas int32 = 1
	dc := cc.AppsV1beta1().Deployments("foo")
	deployment := &appsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "needle",
			Labels: map[string]string{
				"app": "needle",
			},
		},
		Spec: appsv1beta1.DeploymentSpec{
			Replicas: &replicas,
			Strategy: appsv1beta1.DeploymentStrategy{
				RollingUpdate: &appsv1beta1.RollingUpdateDeployment{
					MaxSurge:       IntOrStringPtr(intstr.FromInt(1)),
					MaxUnavailable: IntOrStringPtr(intstr.FromInt(1)),
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "needle",
					},
				},
				Spec: apiv1.PodSpec{
					DNSPolicy: apiv1.DNSClusterFirst,
					Volumes: []apiv1.Volume{
						{
							Name: "config",
							VolumeSource: apiv1.VolumeSource{
								ConfigMap: &apiv1.ConfigMapVolumeSource{
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: "needle-config",
									},
								},
							},
						},
					},
					Containers: []apiv1.Container{
						{
							Name:            "needle",
							Image:           "soon/needle:0.0.1",
							ImagePullPolicy: apiv1.PullIfNotPresent,
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "config",
									MountPath: "/etc/compass",
									ReadOnly:  true,
								},
							},
							Ports: []apiv1.ContainerPort{
								{
									Name:          "grpc",
									ContainerPort: 5000,
								},
							},
							Resources: apiv1.ResourceRequirements{
								Limits: apiv1.ResourceList{
									apiv1.ResourceLimitsCPU:    *resource.NewQuantity(0, resource.DecimalSI),
									apiv1.ResourceLimitsMemory: *resource.NewScaledQuantity(64, resource.Mega),
								},
								Requests: apiv1.ResourceList{
									apiv1.ResourceLimitsCPU:    *resource.NewQuantity(0, resource.DecimalSI),
									apiv1.ResourceLimitsMemory: *resource.NewScaledQuantity(32, resource.Mega),
								},
							},
						},
						{
							Name:            "postgresql",
							Image:           "postgres:10.1-alpine",
							ImagePullPolicy: apiv1.PullIfNotPresent,
							VolumeMounts:    []apiv1.VolumeMount{},
							Env: []apiv1.EnvVar{
								{
									Name:  "POSTGRES_DB",
									Value: "postgres",
								},
								{
									Name:  "POSTGRES_USER",
									Value: "postgres",
								},
								{
									Name:  "POSTGRES_PASS",
									Value: "postgres",
								},
								{
									Name:  "POSTGRES_PGDATA",
									Value: "",
								},
							},
							Ports: []apiv1.ContainerPort{
								{
									Name:          "grpc",
									ContainerPort: 5000,
								},
							},
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceLimitsCPU:    *resource.NewQuantity(100, resource.DecimalSI),
									apiv1.ResourceLimitsMemory: *resource.NewScaledQuantity(256, resource.Mega),
								},
							},
						},
					},
				},
			},
		},
	}

	dc.Create(deployment)

	return nil
}
