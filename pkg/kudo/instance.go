package kudo

import (
	"context"
	"errors"
	"fmt"
	"time"

	kudov1beta1 "github.com/kudobuilder/kudo/pkg/apis/kudo/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachinerytypes "k8s.io/apimachinery/pkg/types"

	"github.com/kudobuilder/test-tools/pkg/client"
)

// Instance wraps a KUDO instance.
type Instance struct {
	kudov1beta1.Instance

	lastPlanCheckUID apimachinerytypes.UID

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
		return Instance{}, fmt.Errorf("failed to get Instance %s in namespace %s: %w", name, namespace, err)
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
		return nil, fmt.Errorf("failed to list Instances in namespace %s: %w", namespace, err)
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

func currentPlanStatusUID(instance Instance, plan string) apimachinerytypes.UID {
	if _, ok := instance.Status.PlanStatus[plan]; !ok {
		// The plan may not have been in use before.
		// We continue, assuming that the plan name is valid and present in OperatorVersion.
		return ""
	}

	ps := instance.Status.PlanStatus[plan]

	return ps.UID
}

func currentPlanStatusAndMessage(instance Instance, plan string) (kudov1beta1.ExecutionStatus, string) {
	if _, ok := instance.Status.PlanStatus[plan]; !ok {
		// The plan may not have been in use before.
		// We continue, assuming that the plan name is valid and present in OperatorVersion.
		return kudov1beta1.ExecutionNeverRun, ""
	}

	// The detailed status message is only availble in the deepest level of the status, so
	// we need to iterate until we fine it.
	ps := instance.Status.PlanStatus[plan]
	if ps.Status != kudov1beta1.ExecutionFatalError || ps.Message != "" {
		return ps.Status, ps.Message
	}

	for _, phaseStatus := range ps.Phases {
		if phaseStatus.Status == kudov1beta1.ExecutionFatalError {
			if phaseStatus.Message != "" {
				return ps.Status, phaseStatus.Message
			}

			for _, stepStatus := range phaseStatus.Steps {
				if stepStatus.Status == kudov1beta1.ExecutionFatalError {
					return ps.Status, stepStatus.Message
				}
			}
		}
	}

	return instance.Status.PlanStatus[plan].Status, instance.Status.PlanStatus[plan].Message
}

// WaitForPlanStatus waits for an instance plan status to reach a status.
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
			err := ctx.Err()
			if errors.Is(err, context.DeadlineExceeded) {
				currentStatus, message := currentPlanStatusAndMessage(instance, plan)

				return PlanStatusTimeout{
					Plan:           plan,
					ExpectedStatus: status,
					ActualStatus:   currentStatus,
					Message:        message,
				}
			}

			return ctx.Err()
		case <-ticker.C:
			if err := instance.Update(); err != nil {
				return err
			}

			activePlanUID := currentPlanStatusUID(instance, plan)
			if activePlanUID == instance.lastPlanCheckUID {
				continue
			}

			currentStatus, _ := currentPlanStatusAndMessage(instance, plan)

			if currentStatus == status {
				instance.lastPlanCheckUID = activePlanUID
				return nil
			}
		}
	}
}

// WaitConfig is used to configure instance wait calls.
type WaitConfig struct {
	Timeout time.Duration
}

// WaitForPlanInProgress waits for an instance plan status to be in progress.
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

// WaitForPlanComplete waits up to 5 minutes for an instance plan status to be completed.
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
		return fmt.Errorf("failed to update Instance %s in namespace %s: %w", instance.Name, instance.Namespace, err)
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
		return fmt.Errorf(
			"failed to update parameters of Instance %s in namespace %s: %w",
			instance.Name,
			instance.Namespace,
			err)
	}

	instance.Instance = *updated

	return nil
}
