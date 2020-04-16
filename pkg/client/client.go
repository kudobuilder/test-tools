// Package client implements wrappers for Kubernetes and KUDO interfaces.
package client

import (
	kudo "github.com/kudobuilder/kudo/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps Kubernetes and KUDO APIs.
type Client struct {
	Kubernetes kubernetes.Interface
	Kudo       kudo.Interface
	Config     rest.Config
	// may be empty in the "in cluster" case
	KubeConfigPath string
}

// NewForConfig creates a Client using a kubeconfig path.
func NewForConfig(kubeconfigPath string) (Client, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return Client{}, err
	}

	kubernetesClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return Client{}, err
	}

	kudoClient, err := kudo.NewForConfig(config)
	if err != nil {
		return Client{}, err
	}

	return Client{
		Kubernetes:     kubernetesClient,
		Kudo:           kudoClient,
		Config:         *config,
		KubeConfigPath: kubeconfigPath,
	}, nil
}

// NewInCluster creates a Client using the service account Kubernetes gives to pods.
func NewInCluster() (Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return Client{}, err
	}

	kubernetesClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return Client{}, err
	}

	kudoClient, err := kudo.NewForConfig(config)
	if err != nil {
		return Client{}, err
	}

	return Client{
		Kubernetes: kubernetesClient,
		Kudo:       kudoClient,
		Config:     *config,
	}, nil
}
