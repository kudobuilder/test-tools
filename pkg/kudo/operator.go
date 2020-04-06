package kudo

import (
	"fmt"
	"log"

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
		false)
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
	return operator.uninstallWithWaitPolicy(noWait)
}

// UninstallWaitForDeletion is the same as Uninstall but waits for the KUDO resources to disappear.
func (operator Operator) UninstallWaitForDeletion() error {
	return operator.uninstallWithWaitPolicy(waitForDisappearance)
}

type waitPolicy int

const (
	noWait waitPolicy = iota
	waitForDisappearance
)

func (operator Operator) uninstallWithWaitPolicy(policy waitPolicy) error {
	if operator.client.Kudo == nil {
		return fmt.Errorf("operator is not initialized")
	}

	options := metav1.DeleteOptions{}

	if policy == waitForDisappearance {
		propagationPolicy := metav1.DeletePropagationForeground
		options.PropagationPolicy = &propagationPolicy
	}

	err := operator.client.Kudo.
		KudoV1beta1().
		Instances(operator.Instance.Namespace).
		Delete(operator.Instance.Name, &options)
	if err != nil {
		return fmt.Errorf(
			"failed to delete Instance %s in namespace %s: %v",
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

	if policy == waitForDisappearance {
		if err := waitForInstanceDeletion(operator); err != nil {
			return err
		}

		if err := waitForOperatorVersionDeletion(operator); err != nil {
			return err
		}

		if err := waitForOperatorDeletion(operator); err != nil {
			return err
		}
	}

	return nil
}

func waitForInstanceDeletion(operator Operator) error {
	objectMeta := operator.Instance.ObjectMeta
	listOptions := getWatchListOptions(objectMeta)

	log.Printf("watching instance disappearance with %#v", listOptions)

	w, err := operator.client.Kudo.KudoV1beta1().Instances(objectMeta.Namespace).Watch(listOptions)
	if err != nil {
		return fmt.Errorf("TODO: %v", err)
	}

	return watchForDeletion(objectMeta, w)
}

func waitForOperatorVersionDeletion(operator Operator) error {
	objectMeta := operator.OperatorVersion.ObjectMeta
	listOptions := getWatchListOptions(objectMeta)

	log.Printf("watching operator version disappearance with %#v", listOptions)

	w, err := operator.client.Kudo.KudoV1beta1().OperatorVersions(objectMeta.Namespace).Watch(listOptions)
	if err != nil {
		return fmt.Errorf("TODO: %v", err)
	}

	return watchForDeletion(objectMeta, w)
}

func waitForOperatorDeletion(operator Operator) error {
	objectMeta := operator.Operator.ObjectMeta
	listOptions := getWatchListOptions(objectMeta)

	log.Printf("watching operator version disappearance with %#v", listOptions)

	w, err := operator.client.Kudo.KudoV1beta1().Operators(objectMeta.Namespace).Watch(listOptions)
	if err != nil {
		return fmt.Errorf("TODO: %v", err)
	}

	return watchForDeletion(objectMeta, w)
}

func getWatchListOptions(objectMeta metav1.ObjectMeta) metav1.ListOptions {
	return metav1.ListOptions{
		ResourceVersion: objectMeta.ResourceVersion,
		LabelSelector:   labels.SelectorFromSet(objectMeta.Labels).String(),
	}
}

func watchForDeletion(objectMeta metav1.ObjectMeta, w watch.Interface) error {
	defer w.Stop()

	for event := range w.ResultChan() {
		log.Printf("got event %#v", event)

		o, err := runtime.DefaultUnstructuredConverter.ToUnstructured(event.Object)
		if err != nil {
			return fmt.Errorf("TODO2: %v", err)
		}

		if event.Type == watch.Deleted && o["metadata"].(map[string]interface{})["name"].(string) == objectMeta.Name {
			break
		}
	}

	return nil
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
