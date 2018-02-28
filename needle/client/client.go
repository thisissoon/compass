package client

import (
	"context"
	"fmt"
	"time"

	"compass/k8s"
	"compass/k8s/tunnel"

	needle "compass/proto/needle/v1"

	"google.golang.org/grpc"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var needlePodLabels labels.Set = labels.Set{"app": "needle"}

// forward opens a port forward to needle
func forward(context string) (*tunnel.Tunnel, error) {
	config, client, err := getKubeClient(context)
	if err != nil {
		return nil, err
	}
	pod, err := getNeedlePodName(client.CoreV1(), "paralympics")
	if err != nil {
		return nil, err
	}
	t := tunnel.New(
		client.CoreV1().RESTClient(),
		config,
		tunnel.WithNamespace("paralympics"),
		tunnel.WithPodName(pod))
	if err := t.Open(); err != nil {
		return nil, err
	}
	return t, nil
}

// configForContext creates a Kubernetes REST client configuration for a given kubeconfig context.
func configForContext(context string) (*rest.Config, error) {
	config, err := k8s.GetConfig(context).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("could not get Kubernetes config for context %q: %s", context, err)
	}
	return config, nil
}

// getKubeClient creates a Kubernetes config and client for a given kubeconfig context.
func getKubeClient(context string) (*rest.Config, kubernetes.Interface, error) {
	config, err := configForContext(context)
	if err != nil {
		return nil, nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get Kubernetes client: %s", err)
	}
	return config, client, nil
}

func getNeedlePodName(client corev1.PodsGetter, namespace string) (string, error) {
	selector := needlePodLabels.AsSelector()
	pod, err := getFirstRunningPod(client, namespace, selector)
	if err != nil {
		return "", err
	}
	return pod.ObjectMeta.GetName(), nil
}

func getFirstRunningPod(client corev1.PodsGetter, namespace string, selector labels.Selector) (*v1.Pod, error) {
	options := metav1.ListOptions{LabelSelector: selector.String()}
	pods, err := client.Pods(namespace).List(options)
	if err != nil {
		return nil, err
	}
	if len(pods.Items) < 1 {
		return nil, fmt.Errorf("could not find needle")
	}
	for _, p := range pods.Items {
		if isPodReady(&p) {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("could not find a ready needle pod")
}

// isPodReady returns true if a pod is ready; false otherwise.
func isPodReady(pod *v1.Pod) bool {
	return isPodReadyConditionTrue(pod.Status)
}

// isPodReady retruns true if a pod is ready; false otherwise.
func isPodReadyConditionTrue(status v1.PodStatus) bool {
	condition := getPodReadyCondition(status)
	return condition != nil && condition.Status == v1.ConditionTrue
}

// getPodReadyCondition extracts the pod ready condition from the given status and returns that.
// Returns nil if the condition is not present.
func getPodReadyCondition(status v1.PodStatus) *v1.PodCondition {
	_, condition := getPodCondition(&status, v1.PodReady)
	return condition
}

// getPodCondition extracts the provided condition from the given status and returns that.
// Returns nil and -1 if the condition is not present, and the index of the located condition.
func getPodCondition(status *v1.PodStatus, conditionType v1.PodConditionType) (int, *v1.PodCondition) {
	if status == nil {
		return -1, nil
	}
	for i := range status.Conditions {
		if status.Conditions[i].Type == conditionType {
			return i, &status.Conditions[i]
		}
	}
	return -1, nil
}

type Client struct {
	tunnel *tunnel.Tunnel
	client needle.NeedleServiceClient
}

func New() (*Client, error) {
	tunnel, err := forward("")
	if err != nil {
		return nil, err
	}
	cc, err := grpc.Dial(
		tunnel.LocalAddress(),
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Second*5),
	)
	if err != nil {
		return nil, err
	}
	return &Client{
		tunnel: tunnel,
		client: needle.NewNeedleServiceClient(cc),
	}, nil
}

func (c *Client) Close() {
	c.tunnel.Close()
}

func (c *Client) PutService(name, namespace, description string) (*needle.Service, error) {
	rsp, err := c.client.PutService(
		context.Background(),
		&needle.PutServiceRequest{
			Service: &needle.Service{
				LogicalName: name,
				Namespace:   namespace,
				Description: description,
			},
		})
	return rsp.GetService(), err
}

func (c *Client) PutDentry(dtab, prefix, dst string, priority int32) (string, error) {
	rsp, err := c.client.PutDentry(
		context.Background(),
		&needle.PutDentryRequest{
			Dentry: &needle.Dentry{
				Dtab:        dtab,
				Prefix:      prefix,
				Destination: dst,
				Priority:    priority,
			},
		})
	if err != nil {
		return "", err
	}
	return rsp.GetDentry().GetId(), nil
}

func (c *Client) DeleteDentryById(id string) (bool, error) {
	rsp, err := c.client.DeleteDentryById(
		context.Background(),
		&needle.DeleteDentryByIdRequest{
			Id: id,
		})
	if err != nil {
		return false, err
	}
	return rsp.GetDeleted(), nil
}

func (c *Client) DeleteDentryByPrefix(dtab, prefix string) (bool, error) {
	rsp, err := c.client.DeleteDentryByPrefix(
		context.Background(),
		&needle.DeleteDentryByPrefixRequest{
			Prefix: prefix,
			Dtab:   dtab,
		})
	if err != nil {
		return false, err
	}
	return rsp.GetDeleted(), nil
}

func (c *Client) RouteToVersion(name, version string) error {
	_, err := c.client.RouteToVersion(
		context.Background(),
		&needle.RouteToVersionRequest{
			LogicalName: name,
			Version:     version,
		})
	return err
}
