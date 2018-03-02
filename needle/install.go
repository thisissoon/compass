package needle

import (
	"compass/version"
	"fmt"
	"math/rand"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"

	appsv1beta1 "k8s.io/api/apps/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	deploymentName             = "needle"
	pvClaimName                = "needle-postgres-pv"
	postgresPasswordSecretName = "needle-postgres-password"
	postgresPasswordSecretKey  = "postgres-password"
)

var Labels = map[string]string{
	"name": "needle",
	"app":  "compass",
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func Password(length int) string {
	return StringWithCharset(length, charset)
}

func IntOrStringPtr(i intstr.IntOrString) *intstr.IntOrString { return &i }

func postgresPasswordSecret() *apiv1.Secret {
	return &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   postgresPasswordSecretName,
			Labels: Labels,
		},
		Type: apiv1.SecretTypeOpaque,
		Data: map[string][]byte{
			postgresPasswordSecretKey: []byte(Password(12)),
		},
	}
}

// pv returns a persistent volumn claim for needle postgres
func pv() *apiv1.PersistentVolumeClaim {
	return &apiv1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:   pvClaimName,
			Labels: Labels,
		},
		Spec: apiv1.PersistentVolumeClaimSpec{
			AccessModes: []apiv1.PersistentVolumeAccessMode{
				apiv1.ReadWriteOnce,
			},
			Resources: apiv1.ResourceRequirements{
				Requests: apiv1.ResourceList{
					apiv1.ResourceStorage: *resource.NewScaledQuantity(8, resource.Giga),
				},
			},
		},
	}
}

// returns a delpoyment for needle
func deployment() *appsv1beta1.Deployment {
	var replicas int32 = 1
	return &appsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   deploymentName,
			Labels: Labels,
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
					Labels: Labels,
				},
				Spec: apiv1.PodSpec{
					DNSPolicy: apiv1.DNSClusterFirst,
					Volumes: []apiv1.Volume{
						{
							Name: "pgdata",
							VolumeSource: apiv1.VolumeSource{
								PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvClaimName,
								},
							},
						},
					},
					Containers: []apiv1.Container{
						{
							Name:            "needle",
							Image:           fmt.Sprintf("soon/needle:%s", version.Version()),
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Ports: []apiv1.ContainerPort{
								{
									Name:          "grpc",
									ContainerPort: 5000,
								},
							},
							Env: []apiv1.EnvVar{
								{
									Name: "NEEDLE_PSQL_PASSWORD",
									ValueFrom: &apiv1.EnvVarSource{
										SecretKeyRef: &apiv1.SecretKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: postgresPasswordSecretName,
											},
											Key: postgresPasswordSecretKey,
										},
									},
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
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      pvClaimName,
									MountPath: "/var/lib/postgresql/data/pgdata",
									SubPath:   "postgresql-db",
								},
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "POSTGRES_DB",
									Value: "needle",
								},
								{
									Name:  "POSTGRES_USER",
									Value: "postgres",
								},
								{
									Name: "POSTGRES_PASS",
									ValueFrom: &apiv1.EnvVarSource{
										SecretKeyRef: &apiv1.SecretKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: postgresPasswordSecretName,
											},
											Key: postgresPasswordSecretKey,
										},
									},
								},
								{
									Name:  "POSTGRES_PGDATA",
									Value: "/var/lib/postgresql/data/pgdata",
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
}

func Install(cc *kubernetes.Clientset, namespace string) error {
	// Create Postgres Password as a Secret
	sc := cc.CoreV1().Secrets(namespace)
	sc.Create(postgresPasswordSecret())
	// Create postgres persistent volumn claim
	pvc := cc.CoreV1().PersistentVolumeClaims(namespace)
	pvc.Create(pv())
	// Create needle deployment
	dc := cc.AppsV1beta1().Deployments(namespace)
	dc.Create(deployment())
	return nil
}
