package kubernetes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/kudobuilder/test-tools/pkg/client"
)

func TestNamespace(t *testing.T) {
	client := client.Client{
		Kubernetes: fake.NewSimpleClientset(),
	}

	const namespace = "test"

	_, err := client.Kubernetes.
		CoreV1().
		Namespaces().
		Get(context.TODO(), namespace, metav1.GetOptions{})
	assert.Error(t, err)

	err = DeleteNamespace(client, namespace)
	assert.Error(t, err)

	err = CreateNamespace(client, namespace)
	assert.NoError(t, err)

	_, err = client.Kubernetes.
		CoreV1().
		Namespaces().
		Get(context.TODO(), namespace, metav1.GetOptions{})
	assert.NoError(t, err)

	err = DeleteNamespace(client, namespace)
	assert.NoError(t, err)

	_, err = client.Kubernetes.
		CoreV1().
		Namespaces().
		Get(context.TODO(), namespace, metav1.GetOptions{})
	assert.Error(t, err)
}
