package debug

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/spf13/afero"

	"github.com/kudobuilder/test-tools/pkg/client"
)

const (
	testArtifactsDirectoryVarName = "TEST_ARTIFACTS_DIRECTORY"
)

// for injecting mock implementations in unit tests
type debugDeps struct {
	artifactsDirectoryBase string
	execCommand            func(name string, arg ...string) *exec.Cmd
	now                    func() time.Time
}

// CollectArtifacts collects useful debugging artifacts from a given namespace.
// Should typically be called if CurrentGinkgoTestDescription().Failed, like this:
//   c, err := client.NewForConfig(KubeConfigPath)
//   if err != nil ...
//   debug.CollectArtifacts(c, afero.NewOsFs(), GinkgoWriter, TestNamespace, KubectlPath)
// Note that this function emits encountered errors to the supplied writer, so there is only a need to inspect its
// return value only if the caller wants to take some action in addition to printing the error.
func CollectArtifacts(client client.Client, fs afero.Fs, writer io.Writer, namespace, kubectlPath string) error {
	return debugDeps{
		artifactsDirectoryBase: os.Getenv(testArtifactsDirectoryVarName),
		execCommand:            exec.Command,
		now:                    time.Now,
	}.collectArtifacts(client, fs, writer, namespace, kubectlPath)
}

func (d debugDeps) collectArtifacts(
	client client.Client, fs afero.Fs, writer io.Writer, namespace, kubectlPath string) error {
	err := d.collectNamespacedResources(client, fs, writer, namespace, kubectlPath)
	if err != nil {
		_, _ = fmt.Fprintf(writer, "collection of resources for debugging failed: %v\n", err)
	}

	return err
}

func (d debugDeps) collectNamespacedResources(
	client client.Client, fs afero.Fs, writer io.Writer, namespace, kubectlPath string) error {
	if d.artifactsDirectoryBase == "" {
		return fmt.Errorf("$%s not set", testArtifactsDirectoryVarName)
	}

	_, _ = fmt.Fprintf(writer, "collecting namespaced resources for debugging...\n")

	cmd := d.execCommand(
		kubectlPath,
		"--kubeconfig",
		client.KubeConfigPath,
		"api-resources",
		"--verbs=list",
		"--namespaced=true",
		"-o",
		"name")
	cmd.Stderr = writer

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("fetching API resource types failed: %v", err)
	}

	artifactsDirectory := path.Join(
		d.artifactsDirectoryBase, fmt.Sprintf("%s-%s", namespace, d.now().Format(time.RFC3339)))

	err = fs.MkdirAll(artifactsDirectory, 0777)
	if err != nil {
		return fmt.Errorf("creating %q failed: %v", artifactsDirectory, err)
	}

	resources := strings.Split(string(output), "\n")
	groupedResources := make(map[string][]string)

	for _, resources := range resources {
		if resources == "" {
			continue
		}

		resourcesAndGroup := strings.SplitN(resources, ".", 2)

		var group string

		if len(resourcesAndGroup) == 2 {
			group = resourcesAndGroup[1]
		}

		groupedResources[group] = append(groupedResources[group], resources)
	}

	var wg sync.WaitGroup

	for group, resourcesSlice := range groupedResources {
		sort.Strings(resourcesSlice)

		resources := strings.Join(resourcesSlice, ",")

		var fileName string

		if group == "" {
			fileName = "resources.yaml"
		} else {
			fileName = fmt.Sprintf("resources-%s.yaml", group)
		}

		wg.Add(1)

		go d.collectResources(client, fs, writer, &wg, namespace, fileName, resources, artifactsDirectory, kubectlPath)
	}

	wg.Wait()

	return nil
}

func (d debugDeps) collectResources(
	client client.Client,
	fs afero.Fs,
	writer io.Writer,
	wg *sync.WaitGroup,
	namespace, fileName,
	resourcesNames,
	directoryName,
	kubectlPath string) {
	defer wg.Done()

	outPath := path.Join(directoryName, fileName)

	output, err := fs.Create(outPath)
	if err != nil {
		_, _ = fmt.Fprintf(writer, "creating %q failed: %v\n", outPath, err)
		return
	}

	cmd := d.execCommand(kubectlPath, "--kubeconfig", client.KubeConfigPath, "get", resourcesNames,
		"--namespace", namespace, "--ignore-not-found", "-o", "yaml")
	cmd.Stdout = output
	cmd.Stderr = writer

	err = cmd.Run()
	if err != nil {
		_, _ = fmt.Fprintf(writer, "fetching %s failed: %v\n", resourcesNames, err)
	}

	empty, err := afero.IsEmpty(fs, outPath)
	if err != nil {
		_, _ = fmt.Fprintf(writer, "checking %s for emptiness failed: %v\n", outPath, err)
	}

	if empty {
		err = fs.Remove(outPath)
		if err != nil {
			_, _ = fmt.Fprintf(writer, "removing empty %s failed: %v\n", outPath, err)
		}
	}
}
