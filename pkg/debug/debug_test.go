package debug

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/kudobuilder/test-tools/pkg/client"
)

func TestCollectArtifacts_Disabled(t *testing.T) {
	sb := strings.Builder{}
	err := debugDeps{
		artifactsDirectoryBase: "",
	}.collectArtifacts(client.Client{}, nil, &sb, "", "")
	assert.EqualError(t, err, "$TEST_ARTIFACTS_DIRECTORY not set")
	assert.Equal(t, sb.String(), "collection of resources for debugging failed: $TEST_ARTIFACTS_DIRECTORY not set\n")
}

type testStruct struct {
	name            string
	apiResourcesOut string
	expectedOut     string
	getOut          map[string]string
	getErr          map[string]string
	getExit         map[string]int
	expectedDirs    []string
	expectedFiles   map[string]string
}

func TestCollectArtifacts(t *testing.T) {
	d := debugDeps{
		artifactsDirectoryBase: "/artifacts",
		now:                    func() time.Time { return time.Time{} },
	}
	tests := []testStruct{
		{
			"zero api resources",
			"",
			"collecting namespaced resources for debugging...\n",
			nil,
			nil,
			nil,
			[]string{
				"/artifacts",
				"/artifacts/ns-0001-01-01T00:00:00Z",
			},
			nil,
		},
		{
			"one resource no output",
			"pods",
			"collecting namespaced resources for debugging...\n",
			map[string]string{
				"pods": "",
			},
			nil,
			nil,
			[]string{
				"/artifacts",
				"/artifacts/ns-0001-01-01T00:00:00Z",
			},
			nil,
		},
		{
			"one resource with output",
			"pods",
			"collecting namespaced resources for debugging...\n",
			map[string]string{
				"pods": "pod1\n",
			},
			nil,
			nil,
			[]string{
				"/artifacts",
				"/artifacts/ns-0001-01-01T00:00:00Z",
			},
			map[string]string{
				"/artifacts/ns-0001-01-01T00:00:00Z/resources.yaml": "pod1\n",
			},
		},
		{
			"two resources of same group",
			"pods\ngremlins\n",
			"collecting namespaced resources for debugging...\npeekaboo\n",
			map[string]string{
				"gremlins,pods": "pod1\n",
			},
			map[string]string{
				"gremlins,pods": "peekaboo\n",
			},
			nil,
			[]string{
				"/artifacts",
				"/artifacts/ns-0001-01-01T00:00:00Z",
			},
			map[string]string{
				"/artifacts/ns-0001-01-01T00:00:00Z/resources.yaml": "pod1\n",
			},
		},
		{
			"two groups of resources",
			"pods\nservices\ninstances.kudo.dev\noperatorversions.kudo.dev\n",
			"collecting namespaced resources for debugging...\n",
			map[string]string{
				"pods,services": "stuff\n",
				"instances.kudo.dev,operatorversions.kudo.dev": "kudo\n",
			},
			nil,
			nil,
			[]string{
				"/artifacts",
				"/artifacts/ns-0001-01-01T00:00:00Z",
			},
			map[string]string{
				"/artifacts/ns-0001-01-01T00:00:00Z/resources-kudo.dev.yaml": "kudo\n",
				"/artifacts/ns-0001-01-01T00:00:00Z/resources.yaml":          "stuff\n",
			},
		},
		{
			"two groups of resources of which one fails",
			"pods\ngremlins\ninstances.kudo.dev\noperatorversions.kudo.dev\n",
			"collecting namespaced resources for debugging...\npeekaboo\nfetching gremlins,pods failed: exit status 1\n",
			map[string]string{
				"gremlins,pods": "peek...",
				"instances.kudo.dev,operatorversions.kudo.dev": "kudo\n",
			},
			map[string]string{
				"gremlins,pods": "peekaboo\n",
			},
			map[string]int{
				"gremlins,pods": 1,
			},
			[]string{
				"/artifacts",
				"/artifacts/ns-0001-01-01T00:00:00Z",
			},
			map[string]string{
				"/artifacts/ns-0001-01-01T00:00:00Z/resources-kudo.dev.yaml": "kudo\n",
				"/artifacts/ns-0001-01-01T00:00:00Z/resources.yaml":          "peek...",
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			sb := strings.Builder{}
			d.execCommand = getExecCommand(t, test)

			err := d.collectArtifacts(client.Client{KubeConfigPath: "kube.config"}, fs, &sb, "ns", "kubectl")
			assert.NoError(t, err)

			assert.Equal(t, test.expectedOut, sb.String())
			assert.NoError(t, afero.Walk(fs, "/", func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					if path == "/" {
						return nil
					}
					assert.Contains(t, test.expectedDirs, path, "directory %q is not expected to exist", path)
				} else {
					_, ok := test.expectedFiles[path]
					assert.True(t, ok, "file %q is not expected to exist", path)
				}
				return nil
			}))
			for file, expected := range test.expectedFiles {
				exists, err := afero.Exists(fs, file)
				assert.True(t, exists, "file %q should exist", file)
				assert.NoError(t, err)
				content, err := afero.ReadFile(fs, file)
				if assert.NoError(t, err) {
					assert.Equal(t, expected, string(content), "file %q has unexpected content", file)
				}
			}
		})
	}
}

func getExecCommand(t *testing.T, test testStruct) func(name string, arg ...string) *exec.Cmd {
	return func(name string, arg ...string) *exec.Cmd {
		assert.Equal(t, name, "kubectl")
		assert.Equal(t, arg[0], "--kubeconfig")
		assert.Equal(t, arg[1], "kube.config")
		cmd := exec.Command(os.Args[0], append([]string{name}, arg...)...) // nolint:gosec

		var (
			stdOut string
			stdErr string
			exit   int
		)

		switch arg[2] {
		case "api-resources":
			stdOut = test.apiResourcesOut
		case "get":
			if declaredOut, ok := test.getOut[arg[3]]; ok {
				stdOut = declaredOut
			} else {
				assert.FailNow(t, "internal test error", "GET %s does not have a declared output", arg[3])
			}

			if declaredErr, ok := test.getErr[arg[3]]; ok {
				stdErr = declaredErr
			}

			if declaredExit, ok := test.getExit[arg[3]]; ok {
				exit = declaredExit
			}
		default:
			t.Fatalf("unexpected argument %q", arg[2])
		}

		cmd.Env = []string{"TEST_MODE=kubectl_wrapper", "STDOUT=" + stdOut, "STDERR=" + stdErr, "EXIT=" + strconv.Itoa(exit)}

		return cmd
	}
}

func kubectlWrapper() int {
	fmt.Print(os.Getenv("STDOUT"))
	fmt.Fprint(os.Stderr, os.Getenv("STDERR"))
	exit, _ := strconv.Atoi(os.Getenv("EXIT"))

	return exit
}

func TestMain(m *testing.M) {
	switch os.Getenv("TEST_MODE") {
	case "kubectl_wrapper":
		os.Exit(kubectlWrapper())
	default:
		os.Exit(m.Run())
	}
}
