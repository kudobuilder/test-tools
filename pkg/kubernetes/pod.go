package kubernetes

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/kudobuilder/test-tools/pkg/client"
	"github.com/kudobuilder/test-tools/pkg/cmd"
)

// Pod wraps a Kubernetes Pod.
type Pod struct {
	corev1.Pod

	client client.Client
}

// GetPod gets a pod in a namespace.
func GetPod(client client.Client, name string, namespace string) (Pod, error) {
	options := metav1.GetOptions{}

	pod, err := client.Kubernetes.
		CoreV1().
		Pods(namespace).
		Get(name, options)
	if err != nil {
		return Pod{}, err
	}

	return Pod{
		Pod:    *pod,
		client: client,
	}, nil
}

// ListPods lists all pods in a namespace.
func ListPods(client client.Client, namespace string) ([]Pod, error) {
	options := metav1.ListOptions{}

	podList, err := client.Kubernetes.
		CoreV1().
		Pods(namespace).
		List(options)
	if err != nil {
		return nil, err
	}

	pods := make([]Pod, 0, len(podList.Items))

	for _, item := range podList.Items {
		pods = append(pods, Pod{
			Pod:    item,
			client: client,
		})
	}

	return pods, nil
}

// Logs returns the (current) logs of a pod's container.
func (pod Pod) ContainerLogs(container string) ([]byte, error) {
	options := corev1.PodLogOptions{
		Container: container,
	}

	result := pod.client.Kubernetes.
		CoreV1().
		Pods(pod.Namespace).
		GetLogs(pod.Name, &options).
		Do()

	if result.Error() != nil {
		return []byte{}, result.Error()
	}

	return result.Raw()
}

// Exec runs a command in a pod's container.
func (pod Pod) ContainerExec(container string, command cmd.Builder) error {
	options := corev1.PodExecOptions{
		Container: container,
		Command:   append([]string{command.Command}, command.Arguments...),
		Stdin:     command.Stdin != nil,
		Stdout:    command.Stdout != nil,
		Stderr:    command.Stderr != nil,
		TTY:       false,
	}

	// adapted from https://github.com/kubernetes/kubernetes/blob/master/test/e2e/framework/exec_util.go
	req := pod.client.Kubernetes.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec").
		Param("container", container)
	req.VersionedParams(&options, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(&pod.client.Config, "POST", req.URL())
	if err != nil {
		return err
	}

	return exec.Stream(remotecommand.StreamOptions{
		Stdin:  command.Stdin,
		Stdout: command.Stdout,
		Stderr: command.Stderr,
		Tty:    false,
	})
}

// Update gets the current Pod status.
func (pod *Pod) Update() error {
	options := metav1.GetOptions{}

	update, err := pod.client.Kubernetes.
		CoreV1().
		Pods(pod.Namespace).
		Get(pod.Name, options)
	if err != nil {
		return err
	}

	pod.Pod = *update

	return nil
}
