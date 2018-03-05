package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"compass/k8s"
	"compass/version"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"

	appsv1beta1 "k8s.io/api/apps/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
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

// password character set
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// random seeder
var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

// password generates a random password for postgres
func password() string {
	b := make([]byte, 12)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// IntOrStringPtr converts returns a pointer to a intstr.IntOrString
func IntOrStringPtr(i intstr.IntOrString) *intstr.IntOrString { return &i }

// postgresPasswordSecret creates a secret that stores a random generated password\
// for the postgres server
func postgresPasswordSecret() *apiv1.Secret {
	return &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   postgresPasswordSecretName,
			Labels: Labels,
		},
		Type: apiv1.SecretTypeOpaque,
		Data: map[string][]byte{
			postgresPasswordSecretKey: []byte(password()),
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

// deployment returns a delpoyment for needle
func deployment(namespace string) *appsv1beta1.Deployment {
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
									apiv1.ResourceCPU:    resource.MustParse("0"),
									apiv1.ResourceMemory: resource.MustParse("64Mi"),
								},
								Requests: apiv1.ResourceList{
									apiv1.ResourceCPU:    resource.MustParse("0"),
									apiv1.ResourceMemory: resource.MustParse("32Mi"),
								},
							},
						},
						{
							Name:            "postgresql",
							Image:           "postgres:10.1-alpine",
							ImagePullPolicy: apiv1.PullIfNotPresent,
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "pgdata",
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
									apiv1.ResourceCPU:    resource.MustParse("0"),
									apiv1.ResourceMemory: resource.MustParse("256Mi"),
								},
							},
						},
					},
				},
			},
		},
	}
}

func installCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install needle, compass's server component into a Kubernetes cluster",
		Run: func(cmd *cobra.Command, _ []string) {
			os.Exit(install())
		},
	}
	return cmd
}

func handleK8sError(err error) int {
	se, ok := err.(*k8serrors.StatusError)
	if !ok {
		fmt.Println(fmt.Sprintf("unexpected error: %s", err))
		return 1
	}
	switch se.Status().Reason {
	case metav1.StatusReasonAlreadyExists:
		fmt.Println("OK: Already exists")
		return 0 // this is fine
	default:
		fmt.Println(fmt.Sprintf("unexpected error: %s", err))
	}
	return 1
}

func createSecret(cs *kubernetes.Clientset, ns string) int {
	fmt.Println("Creating secrets for database...")
	sc := cs.CoreV1().Secrets(ns)
	if _, err := sc.Create(postgresPasswordSecret()); err != nil {
		return handleK8sError(err)
	}
	fmt.Println("OK")
	return 0
}

func createPv(cs *kubernetes.Clientset, ns string) int {
	fmt.Println("Creating persisntent volumn claim for database...")
	pvc := cs.CoreV1().PersistentVolumeClaims(ns)
	if _, err := pvc.Create(pv()); err != nil {
		return handleK8sError(err)
	}
	fmt.Println("OK")
	return 0
}

func createDeployment(cs *kubernetes.Clientset, ns string) int {
	fmt.Println("Creating deployment...")
	dc := cs.AppsV1beta1().Deployments(ns)
	if _, err := dc.Create(deployment(ns)); err != nil {
		return handleK8sError(err)
	}
	fmt.Println("OK")
	return 0
}

func install() int {
	namespace := "compass"
	cs, err := k8s.Clientset()
	if err != nil {
		fmt.Println(err)
		return 1
	}
	if code := createSecret(cs, namespace); code > 0 {
		return code
	}
	if code := createPv(cs, namespace); code > 0 {
		return code
	}
	if code := createDeployment(cs, namespace); code > 0 {
		return code
	}
	fmt.Println("Installed")
	return 0
}
