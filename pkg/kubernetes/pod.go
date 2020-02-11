package kubernetes

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/kudobuilder/test-tools/pkg/cmd"
)

//go:generate go run ../../internal/gen -api CoreV1 -type Pod

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
