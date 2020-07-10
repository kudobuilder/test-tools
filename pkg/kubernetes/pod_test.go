package kubernetes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/kudobuilder/test-tools/pkg/client"
)

func TestPod(t *testing.T) {
	const namespace = "test"

	testPod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: namespace,
		},
	}

	client := client.Client{
		Ctx:        context.TODO(),
		Kubernetes: fake.NewSimpleClientset(testPod.DeepCopyObject()),
	}

	pod, err := GetPod(client, testPod.Name, namespace)
	assert.NoError(t, err)
	assert.Equal(t, Pod{
		Pod:    testPod,
		client: client,
	}, pod)

	pods, err := ListPods(client, namespace)
	assert.NoError(t, err)
	assert.Contains(t, pods, pod)
}
