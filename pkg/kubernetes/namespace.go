package kubernetes

import (
	"fmt"

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
		Create(client.Ctx, &namespace, metav1.CreateOptions{})

	if err != nil {
		return fmt.Errorf("failed to create namespace %s: %w", name, err)
	}

	return nil
}

// DeleteNamespace deletes a namespace.
func DeleteNamespace(client client.Client, name string) error {
	options := metav1.DeleteOptions{}

	err := client.Kubernetes.
		CoreV1().
		Namespaces().
		Delete(client.Ctx, name, options)

	if err != nil {
		return fmt.Errorf("failed to delete namespace %s: %w", name, err)
	}

	return nil
}
