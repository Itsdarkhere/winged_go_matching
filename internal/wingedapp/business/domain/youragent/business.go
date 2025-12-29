package youragent

import (
	"errors"
	"fmt"
	"wingedapp/pgtester/internal/util/validationlib"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
)

type Config struct {
}

func (c *Config) Validate() error {
	return validationlib.Validate(c)
}

type Business struct {
	logger       applog.Logger
	cfg          *Config
	prompter     prompter
	storer       storer
	trans        transactor
	actionLogger actionLogger
}

func NewBusiness(
	logger applog.Logger,
	cfg *Config,
	prompter prompter,
	store storer,
	transactor transactor,
	actionLogger actionLogger,
) (*Business, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	if logger == nil {
		return nil, errors.New("nil logger")
	}
	if prompter == nil {
		return nil, errors.New("nil prompter")
	}
	if store == nil {
		return nil, errors.New("nil convo storer")
	}
	if transactor == nil {
		return nil, errors.New("nil trans")
	}
	if actionLogger == nil {
		return nil, errors.New("nil actionLogger")
	}

	return &Business{
		logger:       logger,
		cfg:          cfg,
		prompter:     prompter,
		storer:       store,
		trans:        transactor,
		actionLogger: actionLogger,
	}, nil
}

func (b *Business) SetPrompter(p prompter) {
	b.prompter = p
}

func (b *Business) SetActionLogger(l actionLogger) {
	b.actionLogger = l
}
