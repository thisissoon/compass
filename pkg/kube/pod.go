package kube

import (
	"errors"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// Pod Errors
var (
	ErrPodNotFound = errors.New("pod not found")
	ErrPodNotReady = errors.New("pod not ready")
)

// getFirstRunningPod returns the first ready pod within the given
// namespace and label selectors
func getFirstRunningPod(client corev1.PodsGetter, namespace string, labels labels.Set) (string, error) {
	options := metav1.ListOptions{
		LabelSelector: labels.AsSelector().String(),
	}
	pods, err := client.Pods(namespace).List(options)
	if err != nil {
		return "", err
	}
	if len(pods.Items) < 1 {
		return "", ErrPodNotFound
	}
	for _, p := range pods.Items {
		if isPodReady(&p) {
			return p.Name, nil
		}
	}
	return "", ErrPodNotReady
}

// isPodReady returns true if a pod is ready; false otherwise.
func isPodReady(pod *v1.Pod) bool {
	for _, c := range pod.Status.Conditions {
		if c.Type == v1.PodReady {
			return true
		}
	}
	return false
}
