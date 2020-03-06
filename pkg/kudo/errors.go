package kudo

import (
	"fmt"

	kudov1beta1 "github.com/kudobuilder/kudo/pkg/apis/kudo/v1beta1"
)

// PlanStatusTimeout is the error returned when waiting for a plan status times out.
type PlanStatusTimeout struct {
	Plan           string
	ExpectedStatus kudov1beta1.ExecutionStatus
	ActualStatus   kudov1beta1.ExecutionStatus
	Message        string
}

// Error returns a pretty-printed error string.
func (p PlanStatusTimeout) Error() string {
	return fmt.Sprintf(
		"timed out waiting for plan %s to have %s status; current plan status is %s with message \"%s\"",
		p.Plan,
		p.ExpectedStatus,
		p.ActualStatus,
		p.Message)
}

// Timeout indicates that this is an error describing a timeout.
func (PlanStatusTimeout) Timeout() bool { return true }

// Temporary indicates that this is a temporary error.
func (PlanStatusTimeout) Temporary() bool { return true }
