package utils

import (
	"fmt"
	"testing"
	"time"

	"github.com/zhavkk/order-service/internal/logger"
)

func TestRetryWithBackoff(t *testing.T) {
	t.Parallel()

	operationSuccess := func() error {
		return nil
	}

	operationFail := func() error {
		return fmt.Errorf("operation failed")
	}
	logger.Init("local")

	operationRetrySuccess := func() func() error {
		attempts := 0
		return func() error {
			attempts++
			if attempts < 3 {
				return fmt.Errorf("temporary error")
			}
			return nil
		}
	}()

	tests := []struct {
		name           string
		operation      func() error
		maxRetries     int
		initialBackoff time.Duration
		expectError    bool
	}{
		{
			name:           "Success on first attempt",
			operation:      operationSuccess,
			maxRetries:     3,
			initialBackoff: 100 * time.Millisecond,
			expectError:    false,
		},
		{
			name:           "Fail after max retries",
			operation:      operationFail,
			maxRetries:     3,
			initialBackoff: 100 * time.Millisecond,
			expectError:    true,
		},
		{
			name:           "Success after retries",
			operation:      operationRetrySuccess,
			maxRetries:     5,
			initialBackoff: 100 * time.Millisecond,
			expectError:    false,
		},
		{
			name:           "No retries allowed",
			operation:      operationFail,
			maxRetries:     0,
			initialBackoff: 100 * time.Millisecond,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := RetryWithBackoff(tt.operation, tt.maxRetries, tt.initialBackoff)
			if (err != nil) != tt.expectError {
				t.Errorf("RetryWithBackoff() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}
