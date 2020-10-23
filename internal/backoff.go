package internal

import (
	"time"

	"github.com/cenkalti/backoff/v4"
)

func NewBackOff() backoff.BackOff {
	const maxElapsedTime = 3 * time.Minute
	backOff := backoff.NewExponentialBackOff()
	backOff.MaxElapsedTime = maxElapsedTime
	return backOff
}
