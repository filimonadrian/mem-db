package util

import (
	"context"
	"fmt"
	"time"
)

func Retry(ctx context.Context, f func() error, retryAttempts int, interval time.Duration) error {
	var err error

	for i := 0; i < retryAttempts; i++ {
		err := f()
		if err == nil {
			return nil
		}

		select {
		case <-time.After(interval):

		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return fmt.Errorf("All retries failed: %v", err)
}
