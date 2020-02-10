package tls

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/kudobuilder/test-tools/pkg/client"
	"github.com/kudobuilder/test-tools/pkg/kubernetes"
)

func TestCertSecret(t *testing.T) {
	client := client.Client{
		Kubernetes: fake.NewSimpleClientset(),
	}

	secret, err := CreateCertSecret("test").
		WithNamespace("test").
		WithCommonName("test").
		Do(client)
	assert.NoError(t, err)

	secrets, err := kubernetes.ListSecrets(client, "test")
	assert.NoError(t, err)
	assert.Contains(t, secrets, secret)
}
