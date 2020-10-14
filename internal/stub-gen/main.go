package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template"
)

//nolint:lll
const stubTemplate = `package kubernetes

// Code generated by stub-gen; DO NOT EDIT.

import (
	"fmt"
{{ if eq .API "CoreV1" }}
	corev1 "k8s.io/api/core/v1"{{ else  if eq .API "AppsV1" }}
	appsv1 "k8s.io/api/apps/v1"{{ else  if eq .API "RbacV1" }}
	rbacv1 "k8s.io/api/rbac/v1"{{ end }}
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kudobuilder/test-tools/pkg/client"
)

// {{ .Type }} wraps a Kubernetes {{ .Type }}.
type {{ .Type }} struct {
	{{ .API | toLower }}.{{ .Type }}

	client client.Client
}

// New{{ .Type }} creates a {{ .Type }} from its Kubernetes {{ .Type }}.
func New{{ .Type }}(client client.Client, {{ .Type | toLower }} {{ .API | toLower }}.{{ .Type }}) ({{ .Type }}, error) {
	created{{ .Type }}, err := client.Kubernetes.
		{{ .API }}().
		{{ .Type }}s({{ if .HasNamespace }}{{ .Type | toLower }}.Namespace{{end}}).
		Create(client.Ctx, &{{ .Type | toLower }}, metav1.CreateOptions{})
	if err != nil {
		return {{ .Type }}{}, fmt.Errorf("failed to create {{ .Type | toLower }} %s{{ if .HasNamespace }} in namespace %s{{ end }}: %w", {{ .Type | toLower }}.Name{{ if .HasNamespace }}, {{ .Type | toLower}}.Namespace{{ end }}, err)
	}

	return {{ .Type }}{
		{{ .Type }}: *created{{ .Type }},
		client: client,
	}, nil
}

// Get{{ .Type }} gets a {{ .Type | toLower }}{{ if .HasNamespace }} in a namespace{{ end }}.
func Get{{ .Type }}(client client.Client, name string{{ if .HasNamespace }}, namespace string{{ end }}) ({{ .Type }}, error) {
	options := metav1.GetOptions{}

	{{ .Type | toLower }}, err := client.Kubernetes.
		{{ .API }}().
		{{ .Type }}s({{ if .HasNamespace }}namespace{{ end }}).
		Get(client.Ctx, name, options)
	if err != nil {
		return {{ .Type }}{}, fmt.Errorf("failed to get {{ .Type | toLower }} %s{{ if .HasNamespace }} in namespace %s{{ end }}: %w", name{{ if .HasNamespace }}, namespace{{ end }}, err)
	}

	return {{ .Type }}{
		{{ .Type }}: *{{ .Type | toLower }},
		client: client,
	}, nil
}

// List{{ .Type }}s lists all {{ .Type | toLower}}s{{ if .HasNamespace }} in a namespace{{ end }}.
func List{{.Type}}s(client client.Client{{ if .HasNamespace }}, namespace string{{ end }}) ([]{{ .Type }}, error) {
	options := metav1.ListOptions{}

	list, err := client.Kubernetes.
		{{ .API }}().
		{{ .Type }}s({{ if .HasNamespace }}namespace{{ end }}).
		List(client.Ctx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to list {{ .Type | toLower }}s{{ if .HasNamespace }} in namespace %s{{ end }}: %w", {{ if .HasNamespace }}namespace, {{ end }}err)
	}

	{{ .Type | toLower }}s := make([]{{ .Type }}, 0, len(list.Items))

	for _, item := range list.Items {
		{{ .Type | toLower }}s = append({{ .Type | toLower }}s, {{ .Type }}{
			{{ .Type }}: item,
			client: client,
		})
	}

	return {{ .Type | toLower }}s, nil
}

// Delete deletes a {{ .Type }} from the Kubernetes cluster.
func ({{ .Type | toLower }} {{ .Type }}) Delete() error {
	options := metav1.DeleteOptions{}

	err := {{ .Type | toLower }}.client.Kubernetes.
		{{ .API }}().
		{{ .Type }}s({{ if .HasNamespace }}{{ .Type | toLower }}.Namespace{{ end }}).
		Delete({{ .Type | toLower}}.client.Ctx, {{ .Type | toLower }}.Name, options)
	if err != nil {
		return fmt.Errorf("failed to delete {{ .Type | toLower }} %s{{ if .HasNamespace }} in namespace %s{{ end }}: %w", {{ .Type | toLower }}.Name, {{ if .HasNamespace }}{{ .Type | toLower}}.Namespace, {{ end }}err)
	}

	return nil
}

// Update gets the current {{ .Type }} status.
func ({{ .Type | toLower }} *{{ .Type }}) Update() error {
	options := metav1.GetOptions{}

	update, err := {{ .Type | toLower }}.client.Kubernetes.
		{{ .API }}().
		{{ .Type }}s({{ if .HasNamespace }}{{ .Type | toLower }}.Namespace{{ end }}).
		Get({{ .Type | toLower}}.client.Ctx, {{ .Type | toLower }}.Name, options)
	if err != nil {
		return fmt.Errorf("failed to update {{ .Type | toLower }} %s{{ if .HasNamespace }} in namespace %s{{ end }}: %w", {{ .Type | toLower }}.Name, {{ if .HasNamespace }}{{ .Type | toLower}}.Namespace, {{ end }}err)
	}

	{{ .Type | toLower }}.{{ .Type }} = *update

	return nil
}

// Save saves the current {{ .Type }}.
func ({{ .Type | toLower }} *{{ .Type }}) Save() error {
	update, err := {{ .Type | toLower }}.client.Kubernetes.
		{{ .API }}().
		{{ .Type }}s({{ if .HasNamespace }}{{ .Type | toLower }}.Namespace{{ end }}).
		Update({{ .Type | toLower}}.client.Ctx, &{{ .Type | toLower }}.{{ .Type }}, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to save {{ .Type | toLower }} %s{{ if .HasNamespace }} in namespace %s{{ end }}: %w", {{ .Type | toLower }}.Name, {{ if .HasNamespace }}{{ .Type | toLower}}.Namespace, {{ end }}err)
	}

	{{ .Type | toLower }}.{{ .Type }} = *update

	return nil
}
`

type parameters struct {
	API          string
	Type         string
	HasNamespace bool
}

// stub-gen creates common function for Kubernetes object wrappers.
func main() {
	var parameters parameters

	flag.StringVar(&parameters.API, "api", "", "kubernetes API")
	flag.StringVar(&parameters.Type, "type", "", "type to generate")

	flag.BoolVar(&parameters.HasNamespace, "hasNamespace", true, "type uses namespace")

	flag.Parse()

	funcMap := template.FuncMap{
		"toLower": strings.ToLower,
	}

	tmpl, err := template.New("stub").Funcs(funcMap).Parse(stubTemplate)
	if err != nil {
		panic(err)
	}

	outputName := fmt.Sprintf("%s.generated.go", strings.ToLower(parameters.Type))

	output, err := os.Create(outputName)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := output.Close(); err != nil {
			panic(err)
		}
	}()

	if err := tmpl.Execute(output, parameters); err != nil {
		panic(err)
	}
}
