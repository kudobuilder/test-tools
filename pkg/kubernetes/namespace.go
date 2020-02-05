package kubernetes

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kudobuilder/test-tools/pkg/client"
)

// CreateNamespace creates a namespace.
func CreateNamespace(client client.Client, name string) error {
	namespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	_, err := client.Kubernetes.
		CoreV1().
		Namespaces().
		Create(&namespace)

	return err
}

// DeleteNamespace deletes a namespace.
func DeleteNamespace(client client.Client, name string) error {
	options := metav1.DeleteOptions{}

	err := client.Kubernetes.
		CoreV1().
		Namespaces().
		Delete(name, &options)

	return err
}
