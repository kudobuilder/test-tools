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
//   debug.CollectArtifacts(afero.NewOsFs(), GinkgoWriter, TestNamespace, KubeConfigPath, KubectlPath)
func CollectArtifacts(fs afero.Fs, writer io.Writer, namespace, kubeConfigPath, kubectlPath string) {
	debugDeps{
		artifactsDirectoryBase: os.Getenv(testArtifactsDirectoryVarName),
		execCommand:            exec.Command,
		now:                    time.Now,
	}.collectArtifacts(fs, writer, namespace, kubeConfigPath, kubectlPath)
}

func (d debugDeps) collectArtifacts(fs afero.Fs, writer io.Writer, namespace, kubeConfigPath, kubectlPath string) {
	err := d.collectNamespacedResources(fs, writer, namespace, kubeConfigPath, kubectlPath)
	if err != nil {
		_, _ = fmt.Fprintf(writer, "collection of resources for debugging failed: %v\n", err)
	}
}

func (d debugDeps) collectNamespacedResources(fs afero.Fs, writer io.Writer, namespace, kubeConfigPath, kubectlPath string) error {
	if d.artifactsDirectoryBase == "" {
		return fmt.Errorf("$%s not set", testArtifactsDirectoryVarName)
	}

	_, _ = fmt.Fprintf(writer, "collecting namespaced resources for debugging...\n")

	cmd := d.execCommand(
		kubectlPath,
		"--kubeconfig",
		kubeConfigPath,
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

	artifactsDirectory := path.Join(d.artifactsDirectoryBase, fmt.Sprintf("%s-%s", namespace, d.now().Format(time.RFC3339)))

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

		go d.collectResources(fs, writer, &wg, namespace, fileName, resources,
			artifactsDirectory, kubectlPath, kubeConfigPath)
	}

	wg.Wait()

	return nil
}

func (d debugDeps) collectResources(fs afero.Fs, writer io.Writer, wg *sync.WaitGroup, namespace, fileName, resourcesNames,
	directoryName, kubectlPath, kubeConfigPath string) {
	defer wg.Done()

	outPath := path.Join(directoryName, fileName)

	output, err := fs.Create(outPath)
	if err != nil {
		_, _ = fmt.Fprintf(writer, "creating %q failed: %v\n", outPath, err)
		return
	}

	cmd := d.execCommand(kubectlPath, "--kubeconfig", kubeConfigPath, "get", resourcesNames,
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
