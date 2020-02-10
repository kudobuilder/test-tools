package kubernetes

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kudobuilder/test-tools/pkg/client"
)

//go:generate go run ../../internal/gen -api CoreV1 -type Secret

type SecretBuilder struct {
	Name      string
	Namespace string
	Data      map[string][]byte
}

func CreateSecret(name string) SecretBuilder {
	return SecretBuilder{
		Name: name,
	}
}

func (builder SecretBuilder) WithNamespace(namespace string) SecretBuilder {
	builder.Namespace = namespace

	return builder
}

func (builder SecretBuilder) WithData(data map[string][]byte) SecretBuilder {
	builder.Data = data

	return builder
}

func (builder SecretBuilder) Do(client client.Client) (Secret, error) {
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      builder.Name,
			Namespace: builder.Namespace,
		},

		Data: builder.Data,
	}

	return NewSecret(client, secret)
}
