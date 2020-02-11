package kubernetes

// Code generated by internal/gen; DO NOT EDIT.

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kudobuilder/test-tools/pkg/client"
)


// PersistentVolumeClaim wraps a Kubernetes PersistentVolumeClaim.
type PersistentVolumeClaim struct {
	corev1.PersistentVolumeClaim

	client client.Client
}

// NewPersistentVolumeClaim creates a PersistentVolumeClaim from its Kubernetes PersistentVolumeClaim.
func NewPersistentVolumeClaim(client client.Client, persistentvolumeclaim corev1.PersistentVolumeClaim) (PersistentVolumeClaim, error) {
	createdPersistentVolumeClaim, err := client.Kubernetes.
		CoreV1().
		PersistentVolumeClaims(persistentvolumeclaim.Namespace).
		Create(&persistentvolumeclaim)
	if err != nil {
		return PersistentVolumeClaim{}, err
	}

	return PersistentVolumeClaim{
		PersistentVolumeClaim:    *createdPersistentVolumeClaim,
		client: client,
	}, nil
}

// GetPersistentVolumeClaim gets a persistentvolumeclaim in a namespace.
func GetPersistentVolumeClaim(client client.Client, name string, namespace string) (PersistentVolumeClaim, error) {
	options := metav1.GetOptions{}

	persistentvolumeclaim, err := client.Kubernetes.
		CoreV1().
		PersistentVolumeClaims(namespace).
		Get(name, options)
	if err != nil {
		return PersistentVolumeClaim{}, err
	}

	return PersistentVolumeClaim{
		PersistentVolumeClaim:    *persistentvolumeclaim,
		client: client,
	}, nil
}

// ListPersistentVolumeClaims lists all persistentvolumeclaims in a namespace.
func ListPersistentVolumeClaims(client client.Client, namespace string) ([]PersistentVolumeClaim, error) {
	options := metav1.ListOptions{}

	list, err := client.Kubernetes.
		CoreV1().
		PersistentVolumeClaims(namespace).
		List(options)
	if err != nil {
		return nil, err
	}

	persistentvolumeclaims := make([]PersistentVolumeClaim, 0, len(list.Items))

	for _, item := range list.Items {
		persistentvolumeclaims = append(persistentvolumeclaims, PersistentVolumeClaim{
			PersistentVolumeClaim:    item,
			client: client,
		})
	}

	return persistentvolumeclaims, nil
}

// Delete deletes a PersistentVolumeClaim from the Kubernetes cluster.
func (persistentvolumeclaim PersistentVolumeClaim) Delete() error {
	options := metav1.DeleteOptions{}

	return persistentvolumeclaim.client.Kubernetes.
		CoreV1().
		PersistentVolumeClaims(persistentvolumeclaim.Namespace).
		Delete(persistentvolumeclaim.Name, &options)
}

// Update gets the current PersistentVolumeClaim status.
func (persistentvolumeclaim *PersistentVolumeClaim) Update() error {
	options := metav1.GetOptions{}

	update, err := persistentvolumeclaim.client.Kubernetes.
		CoreV1().
		PersistentVolumeClaims(persistentvolumeclaim.Namespace).
		Get(persistentvolumeclaim.Name, options)
	if err != nil {
		return err
	}

	persistentvolumeclaim.PersistentVolumeClaim = *update

	return nil
}
