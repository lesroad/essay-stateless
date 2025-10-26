package service

import (
	"context"
	"math"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/sirupsen/logrus"
)

// 24（70/3≈23.333，向上取整为24）
// 20（60/3=20.0，向上取整为20）
// 3（5/2=2.5，向上取整为3）
// 2（4/2=2.0，向上取整为2）
func DivideAndRoundUp(numerator, denominator int64) int64 {
	return int64(math.Ceil(float64(numerator) / float64(denominator)))
}

// 23（70/3≈23.333，向下取整为23）
// 20（60/3=20.0，向下取整为20）
// 2（5/2=2.5，向下取整为2）
// 2（4/2=2.0，向下取整为2）
func DivideAndRoundDown(numerator, denominator int64) int64 {
	return int64(math.Floor(float64(numerator) / float64(denominator)))
}

func RetryWithBackoff(ctx context.Context, operation func() error, stepName string) error {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 30 * time.Second
	b.MaxInterval = 10 * time.Second
	b.InitialInterval = 500 * time.Millisecond

	var attempt int
	maxRetries := 3

	return backoff.Retry(func() error {
		attempt++
		if err := operation(); err != nil {
			if attempt >= maxRetries {
				logrus.Errorf("步骤 %s 达到最大重试次数, 最终失败: %v", stepName, err)
				return backoff.Permanent(err)
			}
			return err
		}
		return nil
	}, backoff.WithContext(b, ctx))
}
