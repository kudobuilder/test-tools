package kubernetes

// Code generated by internal/gen; DO NOT EDIT.

import (
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kudobuilder/test-tools/pkg/client"
)


// StatefulSet wraps a Kubernetes StatefulSet.
type StatefulSet struct {
	appsv1.StatefulSet

	client client.Client
}

// NewStatefulSet creates a StatefulSet from its Kubernetes StatefulSet.
func NewStatefulSet(client client.Client, statefulset appsv1.StatefulSet) (StatefulSet, error) {
	createdStatefulSet, err := client.Kubernetes.
		AppsV1().
		StatefulSets(statefulset.Namespace).
		Create(&statefulset)
	if err != nil {
		return StatefulSet{}, err
	}

	return StatefulSet{
		StatefulSet:    *createdStatefulSet,
		client: client,
	}, nil
}

// GetStatefulSet gets a statefulset in a namespace.
func GetStatefulSet(client client.Client, name string, namespace string) (StatefulSet, error) {
	options := metav1.GetOptions{}

	statefulset, err := client.Kubernetes.
		AppsV1().
		StatefulSets(namespace).
		Get(name, options)
	if err != nil {
		return StatefulSet{}, err
	}

	return StatefulSet{
		StatefulSet:    *statefulset,
		client: client,
	}, nil
}

// ListStatefulSets lists all statefulsets in a namespace.
func ListStatefulSets(client client.Client, namespace string) ([]StatefulSet, error) {
	options := metav1.ListOptions{}

	list, err := client.Kubernetes.
		AppsV1().
		StatefulSets(namespace).
		List(options)
	if err != nil {
		return nil, err
	}

	statefulsets := make([]StatefulSet, 0, len(list.Items))

	for _, item := range list.Items {
		statefulsets = append(statefulsets, StatefulSet{
			StatefulSet:    item,
			client: client,
		})
	}

	return statefulsets, nil
}

// Delete deletes a StatefulSet from the Kubernetes cluster.
func (statefulset StatefulSet) Delete() error {
	options := metav1.DeleteOptions{}

	return statefulset.client.Kubernetes.
		AppsV1().
		StatefulSets(statefulset.Namespace).
		Delete(statefulset.Name, &options)
}
