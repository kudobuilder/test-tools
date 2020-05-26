package kudo

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/Masterminds/semver"
	kudov1beta1 "github.com/kudobuilder/kudo/pkg/apis/kudo/v1beta1"
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
		Get(instance, options)
	if err != nil {
		return Operator{}, fmt.Errorf("failed to get Instance %s in namespace %s: %w", instance, namespace, err)
	}

	ov, err := client.Kudo.
		KudoV1beta1().
		OperatorVersions(namespace).
		Get(i.Spec.OperatorVersion.Name, options)
	if err != nil {
		return Operator{}, fmt.Errorf(
			"failed to get OperatorVersion %s in namespace %s: %w", i.Spec.OperatorVersion.Name, namespace, err)
	}

	o, err := client.Kudo.
		KudoV1beta1().
		Operators(namespace).
		Get(ov.Spec.Operator.Name, options)
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

	kudoClient := kudooperator.NewClientFromK8s(client.Kudo)

	err = kudooperator.InstallPackage(
		kudoClient,
		pkg.Resources,
		false,
		builder.Instance,
		builder.Namespace,
		builder.Parameters,
		false,
		time.Duration(0))
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
// Waits up to timeout for the instance to be deleted, and up to 10 seconds for each of OperatorVersion and Operator.
//
// Note that in the past some issues which were not fully understood were observed when using foreground deletion on
// Instances, see https://github.com/kudobuilder/kudo/issues/1071
func (operator Operator) UninstallWaitForDeletion(timeout time.Duration) error {
	if operator.client.Kudo == nil {
		return fmt.Errorf("operator is not initialized")
	}

	options := metav1.DeleteOptions{}

	if timeout != 0 {
		propagationPolicy := metav1.DeletePropagationForeground
		options.PropagationPolicy = &propagationPolicy
	}

	err := operator.client.Kudo.
		KudoV1beta1().
		Instances(operator.Instance.Namespace).
		Delete(operator.Instance.Name, &options)
	if err != nil {
		return fmt.Errorf(
			"failed to delete Instance %s in namespace %s: %w",
			operator.Instance.Name,
			operator.Instance.Namespace,
			err)
	}

	err = operator.client.Kudo.
		KudoV1beta1().
		OperatorVersions(operator.OperatorVersion.Namespace).
		Delete(operator.OperatorVersion.Name, &options)
	if err != nil {
		return fmt.Errorf(
			"failed to delete OperatorVersion %s in namespace %s: %w",
			operator.OperatorVersion.Name,
			operator.OperatorVersion.Namespace,
			err)
	}

	err = operator.client.Kudo.
		KudoV1beta1().
		Operators(operator.Operator.Namespace).
		Delete(operator.Operator.Name, &options)
	if err != nil {
		return fmt.Errorf(
			"failed to delete Operator %s in namespace %s: %w",
			operator.Operator.Name,
			operator.Operator.Namespace,
			err)
	}

	if timeout != 0 {
		kudoClient := operator.client.Kudo.KudoV1beta1()
		i := operator.Instance
		ov := operator.OperatorVersion
		o := operator.Operator
		instanceTimeoutSec := timeout.Round(time.Second).Milliseconds() / 1000

		const (
			// after the instance is gone, these should in theory disappear quickly, since nothing should refer to them
			operatorVersionTimeoutSec = 10
			operatorTimeoutSec        = 10
		)

		err := waitForDeletion(kudoClient.Instances(i.Namespace), i.ObjectMeta, instanceTimeoutSec)
		if err != nil {
			return err
		}

		err = waitForDeletion(kudoClient.OperatorVersions(ov.Namespace), ov.ObjectMeta, operatorVersionTimeoutSec)
		if err != nil {
			return err
		}

		err = waitForDeletion(kudoClient.Operators(o.Namespace), o.ObjectMeta, operatorTimeoutSec)
		if err != nil {
			return err
		}
	}

	return nil
}

type watcher interface {
	Watch(ops metav1.ListOptions) (watch.Interface, error)
}

func waitForDeletion(watcherInterface watcher, objectMeta metav1.ObjectMeta, timeoutSeconds int64) error {
	listOptions := metav1.ListOptions{
		ResourceVersion: objectMeta.ResourceVersion,
		LabelSelector:   labels.SelectorFromSet(objectMeta.Labels).String(),
		TimeoutSeconds:  &timeoutSeconds,
	}

	w, err := watcherInterface.Watch(listOptions)
	if err != nil {
		return fmt.Errorf("starting watch with %#v failed: %v", listOptions, err)
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

	return fmt.Errorf("timed out waiting for deletion after %d seconds", timeoutSeconds)
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

	kudoClient := kudooperator.NewClientFromK8s(operator.client.Kudo)

	err = kudooperator.UpgradeOperatorVersion(
		kudoClient,
		pkg.Resources.OperatorVersion,
		operator.Instance.Name,
		operator.Instance.Namespace,
		builder.Parameters)
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
