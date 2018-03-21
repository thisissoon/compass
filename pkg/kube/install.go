package kube

import (
	"fmt"
	"math/rand"
	"time"

	"compass/pkg/version"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"

	appsv1beta1 "k8s.io/api/apps/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InstallOptions configures the installation of needle into a k8s cluster
type InstallOptions struct {
	Namespace          string
	RBAC               bool
	ServiceAccountName string
	DeoploymentName    string
	Replicas           int32
	Labels             map[string]string
	PGPVClaimName      string
	PGPassword         string
	PGPWSecretName     string
	PGPWSecretKey      string
}

// Default install options - these can be modified using Option functions
var DefaultInstallOptions = InstallOptions{
	Namespace: "kube-system",
	Labels: map[string]string{
		"name": "needle",
		"app":  "compass",
	},
	RBAC:               false,
	ServiceAccountName: "needle",
	DeoploymentName:    "needle",
	Replicas:           1,
	PGPVClaimName:      "needle-postgres-pv",
	PGPassword:         "", // Balnk will auto generate one
	PGPWSecretName:     "needle-postgres-password",
	PGPWSecretKey:      "password",
}

// An Option can configure InstallOptions
type InstallOption func(*InstallOptions)

// WithNamespace sets the namespace needle should be installed into
func WithInstallNamespace(name string) InstallOption {
	return func(opts *InstallOptions) {
		opts.Namespace = name
	}
}

// WithServiceAccount sets the service account name
func WithInastallServiceAccount(name string) InstallOption {
	return func(opts *InstallOptions) {
		opts.ServiceAccountName = name
	}
}

// WithDeploymentName sets the deployment name
func WithInstallDeploymentName(name string) InstallOption {
	return func(opts *InstallOptions) {
		opts.ServiceAccountName = name
	}
}

// WithLabels sets the labels used for the needle install
func WithInstallLabels(labels map[string]string) InstallOption {
	return func(opts *InstallOptions) {
		opts.Labels = labels
	}
}

// WithRBAC enables RBAC roles to be created for needle on install
func WithInstallRBAC() InstallOption {
	return func(opts *InstallOptions) {
		opts.RBAC = true
	}
}

// Install will intall neeedle into a k8s cluster
func Install(client *kubernetes.Clientset, opts ...InstallOption) error {
	var options = DefaultInstallOptions
	for _, opt := range opts {
		opt(&options)
	}
	// Ensure namespace exists
	if err := createNamespace(client, options); err != nil {
		return err
	}
	// install RBAC roles if RBAC is true
	if options.RBAC {
		// Create service account
		if err := createServiceAccount(client, options); err != nil {
			return err
		}
		// Create a role for needle
		if err := createRole(client, options); err != nil {
			return err
		}
		// Create role binding
		if err := createRoleBinding(client, options); err != nil {
			return err
		}
	}
	// Create a persisntent volume claim for PostgreSQL
	if err := createPVClaim(client, options); err != nil {
		return err
	}
	// Create a postgres secret to store the password
	if err := createSecret(client, options); err != nil {
		return err
	}
	// Creeate deployment
	if err := createDeployment(client, options); err != nil {
		return err
	}
	return nil
}

// handleError handle a k8s error, if something already exists thats ok
// else we just bubble up the error
func handleError(err error) error {
	switch e := err.(type) {
	case *k8serrors.StatusError:
		switch e.Status().Reason {
		case metav1.StatusReasonAlreadyExists:
			return nil
		default:
			return err
		}
	default:
		return err
	}
	return nil
}

// createNamespace creates a namespace within a k8s cluster
func createNamespace(c *kubernetes.Clientset, opts InstallOptions) error {
	_, err := c.CoreV1().Namespaces().Create(&apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   opts.Namespace,
			Labels: opts.Labels,
		},
	})
	return handleError(err)
}

// createServiceAccount create a service account
func createServiceAccount(c *kubernetes.Clientset, opts InstallOptions) error {
	_, err := c.CoreV1().ServiceAccounts(opts.Namespace).Create(&apiv1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:   opts.ServiceAccountName,
			Labels: opts.Labels,
		},
	})
	return handleError(err)
}

// createServiceAccount create a service account
func createRole(c *kubernetes.Clientset, opts InstallOptions) error {
	_, err := c.RbacV1beta1().ClusterRoles().Create(&rbacv1beta1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "needle:services:reader",
			Labels: opts.Labels,
		},
		Rules: []rbacv1beta1.PolicyRule{
			{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"services",
				},
				Verbs: []string{
					"get",
					"list",
				},
			},
		},
	})
	return handleError(err)
}

// Creates a role binding between the service account and role
func createRoleBinding(c *kubernetes.Clientset, opts InstallOptions) error {
	var name = "needle:services:reader"
	_, err := c.RbacV1beta1().ClusterRoleBindings().Create(&rbacv1beta1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: opts.Labels,
		},
		Subjects: []rbacv1beta1.Subject{
			{
				Kind:      rbacv1beta1.ServiceAccountKind,
				Name:      opts.ServiceAccountName,
				Namespace: opts.Namespace,
			},
		},
		RoleRef: rbacv1beta1.RoleRef{
			Kind:     "ClusterRole",
			Name:     name,
			APIGroup: "",
		},
	})
	return handleError(err)
}

// createPVClaim creates a persisntent volume claim for postgres
func createPVClaim(c *kubernetes.Clientset, opts InstallOptions) error {
	_, err := c.CoreV1().PersistentVolumeClaims(opts.Namespace).Create(&apiv1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:   opts.PGPVClaimName,
			Labels: opts.Labels,
			Annotations: map[string]string{
				"volume.alpha.kubernetes.io/storage-class": "default",
			},
		},
		Spec: apiv1.PersistentVolumeClaimSpec{
			AccessModes: []apiv1.PersistentVolumeAccessMode{
				apiv1.ReadWriteOnce,
			},
			Resources: apiv1.ResourceRequirements{
				Requests: apiv1.ResourceList{
					apiv1.ResourceStorage: *resource.NewScaledQuantity(5, resource.Giga),
				},
			},
		},
	})
	return handleError(err)
}

// genPassword generates a random genPassword for postgres
func genPassword() string {
	var charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var seed *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 12)
	for i := range b {
		b[i] = charset[seed.Intn(len(charset))]
	}
	return string(b)
}

// cereateSecret creates a postgres secret with either a provided password
// or a auto generated password
func createSecret(c *kubernetes.Clientset, opts InstallOptions) error {
	var password = opts.PGPassword
	if password == "" {
		password = genPassword()
	}
	_, err := c.CoreV1().Secrets(opts.Namespace).Create(&apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   opts.PGPWSecretName,
			Labels: opts.Labels,
		},
		Type: apiv1.SecretTypeOpaque,
		Data: map[string][]byte{
			opts.PGPWSecretKey: []byte(password),
		},
	})
	return handleError(err)
}

// intOrStringPtr converts returns a pointer to a intstr.IntOrString
func intOrStringPtr(i intstr.IntOrString) *intstr.IntOrString { return &i }

// createDeployment creates a deployment for needle
func createDeployment(c *kubernetes.Clientset, opts InstallOptions) error {
	var serviceAccount = ""
	if opts.RBAC {
		serviceAccount = opts.ServiceAccountName
	}
	_, err := c.AppsV1beta1().Deployments(opts.Namespace).Create(&appsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   opts.DeoploymentName,
			Labels: opts.Labels,
		},
		Spec: appsv1beta1.DeploymentSpec{
			Replicas: &opts.Replicas,
			Strategy: appsv1beta1.DeploymentStrategy{
				RollingUpdate: &appsv1beta1.RollingUpdateDeployment{
					MaxSurge:       intOrStringPtr(intstr.FromInt(1)),
					MaxUnavailable: intOrStringPtr(intstr.FromInt(1)),
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: opts.Labels,
				},
				Spec: apiv1.PodSpec{
					ServiceAccountName: serviceAccount,
					DNSPolicy:          apiv1.DNSClusterFirst,
					Volumes: []apiv1.Volume{
						{
							Name: "pgdata",
							VolumeSource: apiv1.VolumeSource{
								PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
									ClaimName: opts.PGPVClaimName,
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
												Name: opts.PGPWSecretName,
											},
											Key: opts.PGPWSecretKey,
										},
									},
								},
							},
							Resources: apiv1.ResourceRequirements{
								Limits: apiv1.ResourceList{
									apiv1.ResourceCPU:    resource.MustParse("0"),
									apiv1.ResourceMemory: resource.MustParse("32Mi"),
								},
								Requests: apiv1.ResourceList{
									apiv1.ResourceCPU:    resource.MustParse("0"),
									apiv1.ResourceMemory: resource.MustParse("16Mi"),
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
									MountPath: "/var/lib/postgresql/data",
									SubPath:   "data",
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
												Name: opts.PGPWSecretName,
											},
											Key: opts.PGPWSecretKey,
										},
									},
								},
								{
									Name:  "POSTGRES_PGDATA",
									Value: "/var/lib/postgresql/data",
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
									apiv1.ResourceMemory: resource.MustParse("128Mi"),
								},
							},
						},
					},
				},
			},
		},
	})
	return handleError(err)
}
