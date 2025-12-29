package setting

import (
	"fmt"
	"wingedapp/pgtester/internal/util/validationlib"
)

type Business struct {
	cfg    *Config
	storer storer
	trans  transactor
}

type Config struct {
	// ThresholdMinimumAge is the minimum age in seconds that a user must be
	// to be able to use the app.
	ThresholdMinimumAge int `json:"threshold_minimum_age" validate:"required"`
}

func (p *Config) Validate() error {
	return validationlib.Validate(p)
}

func NewBusiness(
	cfg *Config,
	storer storer,
	trans transactor,
) (*Business, error) {
	if storer == nil {
		return nil, fmt.Errorf("nil storer")
	}
	if trans == nil {
		return nil, fmt.Errorf("nil transactor")
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validatE: %w", err)
	}

	return &Business{
		storer: storer,
		cfg:    cfg,
		trans:  trans,
	}, nil
}
