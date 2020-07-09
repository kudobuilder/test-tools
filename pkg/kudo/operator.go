package kudo

import (
	"context"
	"fmt"
	"time"

	"github.com/kudobuilder/kudo/pkg/kudoctl/packages/install"
	"github.com/kudobuilder/kudo/pkg/kudoctl/resources/upgrade"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/Masterminds/semver"
	kudov1beta1 "github.com/kudobuilder/kudo/pkg/apis/kudo/v1beta1"
	cmd_install "github.com/kudobuilder/kudo/pkg/kudoctl/cmd/install"
	"github.com/kudobuilder/kudo/pkg/kudoctl/packages/resolver"
	kudooperator "github.com/kudobuilder/kudo/pkg/kudoctl/util/kudo"
	"github.com/kudobuilder/kudo/pkg/kudoctl/util/repo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kudobuilder/test-tools/pkg/client"
)

// Operator wraps the cluster resources that are installed by KUDO for an operator package.
type Operator struct {
	Name            string
	Instance        Instance
	OperatorVersion kudov1beta1.OperatorVersion
	Operator        kudov1beta1.Operator

	client client.Client
}

func newOperator(client client.Client, name string, instance string, namespace string) (Operator, error) {
	options := metav1.GetOptions{}

	i, err := client.Kudo.
		KudoV1beta1().
		Instances(namespace).
		Get(context.TODO(), instance, options)
	if err != nil {
		return Operator{}, fmt.Errorf("failed to get Instance %s in namespace %s: %w", instance, namespace, err)
	}

	ov, err := client.Kudo.
		KudoV1beta1().
		OperatorVersions(namespace).
		Get(context.TODO(), i.Spec.OperatorVersion.Name, options)
	if err != nil {
		return Operator{}, fmt.Errorf(
			"failed to get OperatorVersion %s in namespace %s: %w", i.Spec.OperatorVersion.Name, namespace, err)
	}

	o, err := client.Kudo.
		KudoV1beta1().
		Operators(namespace).
		Get(context.TODO(), ov.Spec.Operator.Name, options)
	if err != nil {
		return Operator{}, fmt.Errorf(
			"failed to get Operator %s in namespace %s: %w", ov.Spec.Operator.Name, namespace, err)
	}

	return Operator{
		Name: name,
		Instance: Instance{
			Instance: *i,
			client:   client,
		},
		OperatorVersion: *ov,
		Operator:        *o,
		client:          client,
	}, nil
}

// OperatorBuilder tracks the options set for an operator.
type OperatorBuilder struct {
	Name            string
	Namespace       string
	Instance        string
	OperatorVersion *semver.Version
	AppVersion      *semver.Version
	Parameters      map[string]string
	Options         cmd_install.Options
}

// InstallOperator installs a KUDO operator package.
// Additional parameters can be added to this call. The installation
// is started by calling 'Do'.
//   operator, err := kudo.InstallOperator("kafka").
//   	WithNamespace("kafka").
//   	WithInstance("kafka-instance").
//   	WithAppVersion("2.4.0").
//   	Do(client)
func InstallOperator(operator string) OperatorBuilder {
	return OperatorBuilder{
		Name: operator,
	}
}

// WithNamespace sets the namespace in which the operator will be installed.
func (builder OperatorBuilder) WithNamespace(namespace string) OperatorBuilder {
	builder.Namespace = namespace

	return builder
}

// WithInstance sets the name of instance that will be used for the operator.
func (builder OperatorBuilder) WithInstance(instance string) OperatorBuilder {
	builder.Instance = instance

	return builder
}

// WithOperatorVersion sets the version of the operator.
func (builder OperatorBuilder) WithOperatorVersion(version semver.Version) OperatorBuilder {
	builder.OperatorVersion = &version

	return builder
}

// WithAppVersion sets the version of the application that is bundled by the operator.
func (builder OperatorBuilder) WithAppVersion(version semver.Version) OperatorBuilder {
	builder.AppVersion = &version

	return builder
}

// WithParameters sets the parameters to use for the operator instance.
func (builder OperatorBuilder) WithParameters(parameters map[string]string) OperatorBuilder {
	builder.Parameters = parameters

	return builder
}

// Do installs the operator on the cluster.
func (builder OperatorBuilder) Do(client client.Client) (Operator, error) {
	repository, err := repo.NewClient(repo.Default)
	if err != nil {
		return Operator{}, fmt.Errorf("failed to create repository client: %w", err)
	}

	r := resolver.New(repository)

	var operatorVersion string
	if builder.OperatorVersion != nil {
		operatorVersion = builder.OperatorVersion.String()
	}

	var appVersion string
	if builder.AppVersion != nil {
		appVersion = builder.AppVersion.String()
	}

	pkg, err := r.Resolve(builder.Name, appVersion, operatorVersion)
	if err != nil {
		return Operator{}, fmt.Errorf("failed to resolve operator %s: %w", builder.Name, err)
	}

	kudoClient := kudooperator.NewClientFromK8s(client.Kudo, client.Kubernetes)

	installOpts := install.Options{
		SkipInstance:    builder.Options.SkipInstance,
		CreateNamespace: builder.Options.CreateNameSpace,
	}

	if builder.Options.Wait {
		waitDuration := time.Duration(builder.Options.WaitTime) * time.Second
		installOpts.Wait = &waitDuration
	}

	err = install.Package(
		kudoClient,
		builder.Instance,
		builder.Namespace,
		*pkg.Resources,
		builder.Parameters,
		r, installOpts)

	if err != nil {
		return Operator{}, fmt.Errorf("failed to install operator %s: %w", builder.Name, err)
	}

	return newOperator(client, builder.Name, builder.Instance, builder.Namespace)
}

// Uninstall removes the cluster resources of an operator.
// This will remove the Instance, OperatorVersion and Operator!
// We assume that this is the intended behavior for most test cases.
// Don't use this for test cases which have multiple Instances for a single OperatorVersion.
func (operator Operator) Uninstall() error {
	return operator.UninstallWaitForDeletion(0)
}

// UninstallWaitForDeletion is the same as Uninstall but
// initiates a foreground deletion, and waits for the KUDO resources to disappear.
// Waits up to timeout for the instance, operatorversion and operator to be deleted.
// Delete and Wait for I, OV and O has to be done in order as otherwise the OV may end up
// deleted before the Instance is deleted.
//
// Note that in the past some issues which were not fully understood were observed when using foreground deletion on
// Instances, see https://github.com/kudobuilder/kudo/issues/1071
func (operator Operator) UninstallWaitForDeletion(timeout time.Duration) error {
	if operator.client.Kudo == nil {
		return fmt.Errorf("operator is not initialized")
	}

	options := metav1.DeleteOptions{}
	kudoClient := operator.client.Kudo.KudoV1beta1()
	timeoutSec := timeout.Round(time.Second).Milliseconds() / 1000

	if timeout != 0 {
		propagationPolicy := metav1.DeletePropagationForeground
		options.PropagationPolicy = &propagationPolicy
	}

	err := operator.client.Kudo.
		KudoV1beta1().
		Instances(operator.Instance.Namespace).
		Delete(context.TODO(), operator.Instance.Name, options)
	if err != nil {
		return fmt.Errorf(
			"failed to delete Instance %s in namespace %s: %w",
			operator.Instance.Name,
			operator.Instance.Namespace,
			err)
	}
	if timeout != 0 {
		i := operator.Instance
		err := waitForDeletion(kudoClient.Instances(i.Namespace), i.ObjectMeta, timeoutSec)
		if err != nil {
			return err
		}
	}

	err = operator.client.Kudo.
		KudoV1beta1().
		OperatorVersions(operator.OperatorVersion.Namespace).
		Delete(context.TODO(), operator.OperatorVersion.Name, options)
	if err != nil {
		return fmt.Errorf(
			"failed to delete OperatorVersion %s in namespace %s: %w",
			operator.OperatorVersion.Name,
			operator.OperatorVersion.Namespace,
			err)
	}
	if timeout != 0 {
		ov := operator.OperatorVersion
		err = waitForDeletion(kudoClient.OperatorVersions(ov.Namespace), ov.ObjectMeta, timeoutSec)
		if err != nil {
			return err
		}
	}

	err = operator.client.Kudo.
		KudoV1beta1().
		Operators(operator.Operator.Namespace).
		Delete(context.TODO(), operator.Operator.Name, options)
	if err != nil {
		return fmt.Errorf(
			"failed to delete Operator %s in namespace %s: %w",
			operator.Operator.Name,
			operator.Operator.Namespace,
			err)
	}
	if timeout != 0 {
		o := operator.Operator
		err = waitForDeletion(kudoClient.Operators(o.Namespace), o.ObjectMeta, timeoutSec)
		if err != nil {
			return err
		}
	}

	return nil
}

type watcher interface {
	Watch(ctx context.Context, ops metav1.ListOptions) (watch.Interface, error)
}

func waitForDeletion(watcherInterface watcher, objectMeta metav1.ObjectMeta, timeoutSeconds int64) error {
	listOptions := metav1.ListOptions{
		ResourceVersion: objectMeta.ResourceVersion,
		LabelSelector:   labels.SelectorFromSet(objectMeta.Labels).String(),
		TimeoutSeconds:  &timeoutSeconds,
	}

	w, err := watcherInterface.Watch(context.TODO(), listOptions)
	if err != nil {
		return fmt.Errorf("starting watch of %s/%s with %#v failed: %v", objectMeta.Namespace, objectMeta.Name, listOptions, err)
	}

	defer w.Stop()

	for event := range w.ResultChan() {
		o, err := runtime.DefaultUnstructuredConverter.ToUnstructured(event.Object)
		if err != nil {
			return fmt.Errorf("converting object event %#v to unstructured failed: %v", event.Object, err)
		}

		if event.Type != watch.Deleted {
			continue
		}

		if o["metadata"].(map[string]interface{})["name"].(string) == objectMeta.Name {
			return nil
		}
	}

	return fmt.Errorf("timed out waiting for deletion of %s/%s after %d seconds", objectMeta.Namespace, objectMeta.Name, timeoutSeconds)
}

// UpgradeBuilder tracks the options set for an upgrade.
type UpgradeBuilder struct {
	Name            string
	OperatorVersion *semver.Version
	AppVersion      *semver.Version
	Parameters      map[string]string
}

// WithOperator sets the name of the operator to upgrade with.
func (builder UpgradeBuilder) WithOperator(name string) UpgradeBuilder {
	builder.Name = name

	return builder
}

// ToOperatorVersion sets the operator version to upgrade to.
func (builder UpgradeBuilder) ToOperatorVersion(version semver.Version) UpgradeBuilder {
	builder.OperatorVersion = &version

	return builder
}

// ToAppVersion sets the application version to upgrade to.
func (builder UpgradeBuilder) ToAppVersion(version semver.Version) UpgradeBuilder {
	builder.AppVersion = &version

	return builder
}

// WithParameters sets or changes parameters that the upgraded operator should use.
func (builder UpgradeBuilder) WithParameters(parameters map[string]string) UpgradeBuilder {
	builder.Parameters = parameters

	return builder
}

// Do upgrades the operator.
func (builder UpgradeBuilder) Do(operator *Operator) error {
	repository, err := repo.NewClient(repo.Default)
	if err != nil {
		return fmt.Errorf("failed to create repository client: %w", err)
	}

	r := resolver.New(repository)

	name := operator.Name
	if builder.Name != "" {
		name = builder.Name
	}

	var operatorVersion string
	if builder.OperatorVersion != nil {
		operatorVersion = builder.OperatorVersion.String()
	}

	var appVersion string
	if builder.AppVersion != nil {
		appVersion = builder.AppVersion.String()
	}

	pkg, err := r.Resolve(name, appVersion, operatorVersion)
	if err != nil {
		return fmt.Errorf("failed to resolve operator %s: %w", name, err)
	}

	kudoClient := kudooperator.NewClientFromK8s(operator.client.Kudo, operator.client.Kubernetes)

	err = upgrade.OperatorVersion(
		kudoClient,
		pkg.Resources.OperatorVersion,
		operator.Instance.Name,
		builder.Parameters,
		resolver.New(repository))

	if err != nil {
		return fmt.Errorf(
			"failed to upgrade OperatorVersion for Instance %s in namespace %s: %w",
			operator.Instance.Name,
			operator.Instance.Namespace,
			err)
	}

	newOperator, err := newOperator(
		operator.client,
		operator.Name,
		operator.Instance.Name,
		operator.Instance.Namespace)
	if err != nil {
		return err
	}

	operator = &newOperator

	return nil
}

// UpgradeOperator upgrades an operator.
// Additional options can be added to this call. The upgrade is started by calling 'Do'.
//   kudo.UpgradeOperator().
//   	ToAppVersion("1.0.1").
//   	Do(operator)
func UpgradeOperator() UpgradeBuilder {
	return UpgradeBuilder{}
}
