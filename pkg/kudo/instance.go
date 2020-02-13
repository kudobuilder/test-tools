package kudo

import (
	"context"
	"time"

	kudov1beta1 "github.com/kudobuilder/kudo/pkg/apis/kudo/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kudobuilder/test-tools/pkg/client"
)

// Instance wraps a KUDO instance.
type Instance struct {
	kudov1beta1.Instance

	client client.Client
}

// GetInstance retrieves a KUDO instance in a namespace.
func GetInstance(client client.Client, name string, namespace string) (Instance, error) {
	options := metav1.GetOptions{}

	instance, err := client.Kudo.
		KudoV1beta1().
		Instances(namespace).
		Get(name, options)
	if err != nil {
		return Instance{}, err
	}

	return Instance{
		Instance: *instance,
		client:   client,
	}, nil
}

// ListInstances lists all KUDO instances in a namespace.
func ListInstances(client client.Client, namespace string) ([]Instance, error) {
	options := metav1.ListOptions{}

	instanceList, err := client.Kudo.
		KudoV1beta1().
		Instances(namespace).
		List(options)
	if err != nil {
		return nil, err
	}

	instances := make([]Instance, 0, len(instanceList.Items))

	for _, item := range instanceList.Items {
		instances = append(instances, Instance{
			Instance: item,
			client:   client,
		})
	}

	return instances, nil
}

// WaitForStatus waits for an instance plan status to reach a status.
// A ticker polls the current instance status until the desired status is reached for a specific plan.
// A context can abort the polling.
func (instance Instance) WaitForPlanStatus(
	ctx context.Context,
	ticker *time.Ticker,
	plan string,
	status kudov1beta1.ExecutionStatus) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := instance.Update(); err != nil {
				return err
			}

			if _, ok := instance.Status.PlanStatus[plan]; !ok {
				// The plan may not have been in use before.
				// We continue, assuming that the plan name is valid and present in OperatorVersion.
				continue
			}

			planStatus := instance.Status.PlanStatus[plan]
			if planStatus.Status == status {
				return nil
			}
		}
	}
}

// WaitConfig is used to configure instance wait calls.
type WaitConfig struct {
	Timeout time.Duration
}

// WaitForDeployInProgress waits for an instance plan status to be in progress.
// By default it waits for 30 seconds unless overridden with a WaitTimeout.
func (instance Instance) WaitForPlanInProgress(plan string, options ...WaitOption) error {
	config := WaitConfig{
		Timeout: time.Second * 30,
	}

	for _, option := range options {
		option(&config)
	}

	ctx, cancel := context.WithTimeout(context.TODO(), config.Timeout)
	defer cancel()

	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	return instance.WaitForPlanStatus(ctx, ticker, plan, kudov1beta1.ExecutionInProgress)
}

// WaitForDeployComplete waits up to 5 minutes for an instance plan status to be completed.
// By default it waits for 5 minutes unless overridden with a WaitTimeout.
func (instance Instance) WaitForPlanComplete(plan string, options ...WaitOption) error {
	config := WaitConfig{
		Timeout: time.Minute * 5,
	}

	for _, option := range options {
		option(&config)
	}

	ctx, cancel := context.WithTimeout(context.TODO(), config.Timeout)
	defer cancel()

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	return instance.WaitForPlanStatus(ctx, ticker, plan, kudov1beta1.ExecutionComplete)
}

// Update gets the current instance state.
func (instance *Instance) Update() error {
	options := metav1.GetOptions{}

	update, err := instance.client.Kudo.
		KudoV1beta1().
		Instances(instance.Namespace).
		Get(instance.Name, options)
	if err != nil {
		return err
	}

	instance.Instance = *update

	return nil
}

// UpdateParameters merges new parameters with the existing ones.
// The instance will be updated on the server to use the new parameters.
// These updated can trigger plans.
func (instance *Instance) UpdateParameters(parameters map[string]string) error {
	if err := instance.Update(); err != nil {
		return err
	}

	current := instance.Instance

	if current.Spec.Parameters == nil {
		current.Spec.Parameters = make(map[string]string)
	}

	for k, v := range parameters {
		current.Spec.Parameters[k] = v
	}

	updated, err := instance.client.Kudo.
		KudoV1beta1().
		Instances(instance.Namespace).
		Update(&current)
	if err != nil {
		return err
	}

	instance.Instance = *updated

	return nil
}
