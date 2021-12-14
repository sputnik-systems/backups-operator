package v1alpha1

import (
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
)

type ExponentialBackOffSpec struct {
	InitialInterval string `json:"initialInterval,omitempty"`
	MaxInterval     string `json:"maxInterval,omitempty"`
	MaxElapsedTime  string `json:"maxElapsedTime,omitempty"`
}

func (e *ExponentialBackOffSpec) GetBackOff() (*backoff.ExponentialBackOff, error) {
	bo := backoff.NewExponentialBackOff()

	if e != nil {
		if e.InitialInterval != "" {
			d, err := time.ParseDuration(e.InitialInterval)
			if err != nil {
				return nil, fmt.Errorf("failed to parse InitialInterval: %w", err)
			}
			bo.InitialInterval = d
		}
		if e.MaxInterval != "" {
			d, err := time.ParseDuration(e.MaxInterval)
			if err != nil {
				return nil, fmt.Errorf("failed to parse MaxInterval: %w", err)
			}
			bo.MaxInterval = d
		}
		if e.MaxElapsedTime != "" {
			d, err := time.ParseDuration(e.MaxElapsedTime)
			if err != nil {
				return nil, fmt.Errorf("failed to parse MaxElapsedTime: %w", err)
			}
			bo.MaxElapsedTime = d
		}
	}

	return bo, nil
}
