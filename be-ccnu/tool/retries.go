package tool

import (
	"fmt"
	"time"
)

func Retry(attempts int, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		time.Sleep(time.Duration(i) * time.Second)
	}
	return fmt.Errorf("重试 %d 次后仍失败: %w", attempts, err)
}

func MustRetry[T any](fn func() (T, error)) (T, error) {
	var (
		result T
		err    error
	)
	for i := 0; i < 3; i++ {
		result, err = fn()
		if err == nil {
			return result, nil
		}

		time.Sleep(time.Duration(i) * time.Second)
	}

	return result, err
}
