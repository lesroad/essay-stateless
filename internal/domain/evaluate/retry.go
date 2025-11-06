package evaluate

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries   int
	InitialDelay time.Duration
	MaxDelay     time.Duration
}

// DefaultRetryConfig 默认重试配置
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:   3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     2 * time.Second,
	}
}

// RetryExecutor 重试执行器
type RetryExecutor struct {
	config RetryConfig
}

// NewRetryExecutor 创建重试执行器
func NewRetryExecutor(config RetryConfig) *RetryExecutor {
	return &RetryExecutor{
		config: config,
	}
}

// Execute 执行带重试的函数
func (r *RetryExecutor) Execute(ctx context.Context, fn func() error, stepName string) error {
	var lastErr error
	delay := r.config.InitialDelay

	for i := 0; i <= r.config.MaxRetries; i++ {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 执行函数
		if err := fn(); err != nil {
			lastErr = err

			if i < r.config.MaxRetries {
				logrus.WithFields(logrus.Fields{
					"step":    stepName,
					"attempt": i + 1,
					"error":   err,
					"delay":   delay / time.Microsecond,
				}).Warn("操作失败，准备重试")

				// 等待后重试
				select {
				case <-time.After(delay):
					// 指数退避
					delay = delay * 2
					if delay > r.config.MaxDelay {
						delay = r.config.MaxDelay
					}
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		} else {
			// 成功
			return nil
		}
	}

	return fmt.Errorf("%s 失败，已重试 %d 次: %w", stepName, r.config.MaxRetries, lastErr)
}
