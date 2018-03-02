package portforward

import (
	"fmt"

	"compass/k8s"
	"compass/k8s/tunnel"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// needle pod selectors
var needlePodLabels labels.Set = labels.Set{"app": "needle"}

// Options configures portforward options
type Options struct {
	Context   string
	Namespace string
}

// forward opens a port forward tunel to needle
func New(opts Options) (*tunnel.Tunnel, error) {
	config, client, err := getKubeClient(opts.Context)
	if err != nil {
		return nil, err
	}
	pod, err := getNeedlePodName(client.CoreV1(), opts.Namespace)
	if err != nil {
		return nil, err
	}
	t := tunnel.New(
		client.CoreV1().RESTClient(),
		config,
		tunnel.WithNamespace(opts.Namespace),
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
