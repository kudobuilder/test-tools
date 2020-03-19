package kubernetes

// Code generated by stub-gen; DO NOT EDIT.

import (
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kudobuilder/test-tools/pkg/client"
)

// Role wraps a Kubernetes Role.
type Role struct {
	rbacv1.Role

	client client.Client
}

// NewRole creates a Role from its Kubernetes Role.
func NewRole(client client.Client, role rbacv1.Role) (Role, error) {
	createdRole, err := client.Kubernetes.
		RbacV1().
		Roles(role.Namespace).
		Create(&role)
	if err != nil {
		return Role{}, fmt.Errorf("failed to create role %s in namespace %s: %w", role.Name, role.Namespace, err)
	}

	return Role{
		Role:   *createdRole,
		client: client,
	}, nil
}

// GetRole gets a role in a namespace.
func GetRole(client client.Client, name string, namespace string) (Role, error) {
	options := metav1.GetOptions{}

	role, err := client.Kubernetes.
		RbacV1().
		Roles(namespace).
		Get(name, options)
	if err != nil {
		return Role{}, fmt.Errorf("failed to get role %s in namespace %s: %w", name, namespace, err)
	}

	return Role{
		Role:   *role,
		client: client,
	}, nil
}

// ListRoles lists all roles in a namespace.
func ListRoles(client client.Client, namespace string) ([]Role, error) {
	options := metav1.ListOptions{}

	list, err := client.Kubernetes.
		RbacV1().
		Roles(namespace).
		List(options)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles in namespace %s: %w", namespace, err)
	}

	roles := make([]Role, 0, len(list.Items))

	for _, item := range list.Items {
		roles = append(roles, Role{
			Role:   item,
			client: client,
		})
	}

	return roles, nil
}

// Delete deletes a Role from the Kubernetes cluster.
func (role Role) Delete() error {
	options := metav1.DeleteOptions{}

	err := role.client.Kubernetes.
		RbacV1().
		Roles(role.Namespace).
		Delete(role.Name, &options)
	if err != nil {
		return fmt.Errorf("failed to delete role %s in namespace %s: %w", role.Name, role.Namespace, err)
	}

	return nil
}

// Update gets the current Role status.
func (role *Role) Update() error {
	options := metav1.GetOptions{}

	update, err := role.client.Kubernetes.
		RbacV1().
		Roles(role.Namespace).
		Get(role.Name, options)
	if err != nil {
		return fmt.Errorf("failed to update role %s in namespace %s: %w", role.Name, role.Namespace, err)
	}

	role.Role = *update

	return nil
}
