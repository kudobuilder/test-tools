package kubernetes

// Code generated by stub-gen; DO NOT EDIT.

import (
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kudobuilder/test-tools/pkg/client"
)

// RoleBinding wraps a Kubernetes RoleBinding.
type RoleBinding struct {
	rbacv1.RoleBinding

	client client.Client
}

// NewRoleBinding creates a RoleBinding from its Kubernetes RoleBinding.
func NewRoleBinding(client client.Client, rolebinding rbacv1.RoleBinding) (RoleBinding, error) {
	createdRoleBinding, err := client.Kubernetes.
		RbacV1().
		RoleBindings(rolebinding.Namespace).
		Create(&rolebinding)
	if err != nil {
		return RoleBinding{}, fmt.Errorf("failed to create rolebinding %s in namespace %s: %w", rolebinding.Name, rolebinding.Namespace, err)
	}

	return RoleBinding{
		RoleBinding: *createdRoleBinding,
		client:      client,
	}, nil
}

// GetRoleBinding gets a rolebinding in a namespace.
func GetRoleBinding(client client.Client, name string, namespace string) (RoleBinding, error) {
	options := metav1.GetOptions{}

	rolebinding, err := client.Kubernetes.
		RbacV1().
		RoleBindings(namespace).
		Get(name, options)
	if err != nil {
		return RoleBinding{}, fmt.Errorf("failed to get rolebinding %s in namespace %s: %w", name, namespace, err)
	}

	return RoleBinding{
		RoleBinding: *rolebinding,
		client:      client,
	}, nil
}

// ListRoleBindings lists all rolebindings in a namespace.
func ListRoleBindings(client client.Client, namespace string) ([]RoleBinding, error) {
	options := metav1.ListOptions{}

	list, err := client.Kubernetes.
		RbacV1().
		RoleBindings(namespace).
		List(options)
	if err != nil {
		return nil, fmt.Errorf("failed to list rolebindings in namespace %s: %w", namespace, err)
	}

	rolebindings := make([]RoleBinding, 0, len(list.Items))

	for _, item := range list.Items {
		rolebindings = append(rolebindings, RoleBinding{
			RoleBinding: item,
			client:      client,
		})
	}

	return rolebindings, nil
}

// Delete deletes a RoleBinding from the Kubernetes cluster.
func (rolebinding RoleBinding) Delete() error {
	options := metav1.DeleteOptions{}

	err := rolebinding.client.Kubernetes.
		RbacV1().
		RoleBindings(rolebinding.Namespace).
		Delete(rolebinding.Name, &options)
	if err != nil {
		return fmt.Errorf("failed to delete rolebinding %s in namespace %s: %w", rolebinding.Name, rolebinding.Namespace, err)
	}

	return nil
}

// Update gets the current RoleBinding status.
func (rolebinding *RoleBinding) Update() error {
	options := metav1.GetOptions{}

	update, err := rolebinding.client.Kubernetes.
		RbacV1().
		RoleBindings(rolebinding.Namespace).
		Get(rolebinding.Name, options)
	if err != nil {
		return fmt.Errorf("failed to update rolebinding %s in namespace %s: %w", rolebinding.Name, rolebinding.Namespace, err)
	}

	rolebinding.RoleBinding = *update

	return nil
}
