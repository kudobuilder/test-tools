package kubernetes

// Code generated by stub-gen; DO NOT EDIT.

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kudobuilder/test-tools/pkg/client"
)

// Service wraps a Kubernetes Service.
type Service struct {
	corev1.Service

	client client.Client
}

// NewService creates a Service from its Kubernetes Service.
func NewService(client client.Client, service corev1.Service) (Service, error) {
	createdService, err := client.Kubernetes.
		CoreV1().
		Services(service.Namespace).
		Create(&service)
	if err != nil {
		return Service{}, fmt.Errorf("failed to create service %s in namespace %s: %w", service.Name, service.Namespace, err)
	}

	return Service{
		Service: *createdService,
		client:  client,
	}, nil
}

// GetService gets a service in a namespace.
func GetService(client client.Client, name string, namespace string) (Service, error) {
	options := metav1.GetOptions{}

	service, err := client.Kubernetes.
		CoreV1().
		Services(namespace).
		Get(name, options)
	if err != nil {
		return Service{}, fmt.Errorf("failed to get service %s in namespace %s: %w", name, namespace, err)
	}

	return Service{
		Service: *service,
		client:  client,
	}, nil
}

// ListServices lists all services in a namespace.
func ListServices(client client.Client, namespace string) ([]Service, error) {
	options := metav1.ListOptions{}

	list, err := client.Kubernetes.
		CoreV1().
		Services(namespace).
		List(options)
	if err != nil {
		return nil, fmt.Errorf("failed to list services in namespace %s: %w", namespace, err)
	}

	services := make([]Service, 0, len(list.Items))

	for _, item := range list.Items {
		services = append(services, Service{
			Service: item,
			client:  client,
		})
	}

	return services, nil
}

// Delete deletes a Service from the Kubernetes cluster.
func (service Service) Delete() error {
	options := metav1.DeleteOptions{}

	err := service.client.Kubernetes.
		CoreV1().
		Services(service.Namespace).
		Delete(service.Name, &options)
	if err != nil {
		return fmt.Errorf("failed to delete service %s in namespace %s: %w", service.Name, service.Namespace, err)
	}

	return nil
}

// Update gets the current Service status.
func (service *Service) Update() error {
	options := metav1.GetOptions{}

	update, err := service.client.Kubernetes.
		CoreV1().
		Services(service.Namespace).
		Get(service.Name, options)
	if err != nil {
		return fmt.Errorf("failed to update service %s in namespace %s: %w", service.Name, service.Namespace, err)
	}

	service.Service = *update

	return nil
}

// Save saves the current Service.
func (service *Service) Save() error {
	update, err := service.client.Kubernetes.
		CoreV1().
		Services(service.Namespace).
		Update(&service.Service)
	if err != nil {
		return fmt.Errorf("failed to save service %s in namespace %s: %w", service.Name, service.Namespace, err)
	}

	service.Service = *update

	return nil
}
