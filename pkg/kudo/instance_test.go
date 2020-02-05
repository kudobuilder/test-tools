package kudo

import (
	"testing"

	kudov1beta1 "github.com/kudobuilder/kudo/pkg/apis/kudo/v1beta1"
	"github.com/kudobuilder/kudo/pkg/client/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kudobuilder/test-tools/pkg/client"
)

func TestInstances(t *testing.T) {
	const namespace = "test"

	testInstance := kudov1beta1.Instance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-instance",
			Namespace: namespace,
		},
	}

	client := client.Client{
		Kudo: fake.NewSimpleClientset(testInstance.DeepCopyObject()),
	}

	instance, err := GetInstance(client, testInstance.Name, namespace)
	assert.NoError(t, err)
	assert.Equal(t, Instance{
		Instance: testInstance,
		client:   client,
	}, instance)

	instances, err := ListInstances(client, namespace)
	assert.NoError(t, err)
	assert.Contains(t, instances, instance)
}
