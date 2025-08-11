package utils

import (
	"fmt"
	"time"

	"github.com/zhavkk/order-service/internal/logger"
)

// exponential backoff retry mechanism
func RetryWithBackoff(operation func() error, maxRetries int, initialBackoff time.Duration) error {
	backoff := initialBackoff
	for i := 0; i < maxRetries; i++ {
		if err := operation(); err != nil {
			logger.Log.Warn("Retrying operation", "attempt", i+1, "error", err)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}
		return nil
	}
	return fmt.Errorf("operation failed after %d retries", maxRetries)
}
