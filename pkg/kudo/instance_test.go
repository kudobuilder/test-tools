package kudo

import (
	"testing"
	"time"

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

func TestWaitTimeout(t *testing.T) {
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

	deadline := time.Now().Add(time.Second)

	err = instance.WaitForPlanComplete("deploy", WaitTimeout(time.Millisecond*1))
	assert.EqualError(t, err, "timed out waiting for plan deploy to have COMPLETE status; current plan status is NEVER_RUN with message \"\"") //nolint:lll
	assert.True(t, time.Now().Before(deadline))
}
