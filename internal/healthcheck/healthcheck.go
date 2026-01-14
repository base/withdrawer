package healthcheck

import (
	"context"
	"errors"
	"time"
)

// Checker defines a minimal interface for health checks.
type Checker interface {
	Check(ctx context.Context) error
}

// WithdrawalChecker verifies that the withdrawal loop is responsive.
type WithdrawalChecker struct {
	LastSuccessfulRun time.Time
	MaxAllowedDelay   time.Duration
}

// Check returns an error if the last successful withdrawal run
// exceeds the allowed delay threshold.
func (w *WithdrawalChecker) Check(ctx context.Context) error {
	if time.Since(w.LastSuccessfulRun) > w.MaxAllowedDelay {
		return errors.New("withdrawal process is stalled")
	}
	return nil
}
