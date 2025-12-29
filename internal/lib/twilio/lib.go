package twilio

import (
	"context"
	"fmt"
	"github.com/twilio/twilio-go"
	twilioAPI "github.com/twilio/twilio-go/rest/api/v2010"
)

type Lib struct {
	cfg    *Config
	client *twilio.RestClient
}

// New creates a new Twilio client with the provided configuration.
func New(cfg *Config) (*Lib, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	l := &Lib{
		cfg: cfg,
		client: twilio.NewRestClientWithParams(twilio.ClientParams{
			Username: cfg.AccountSID,
			Password: cfg.AuthToken,
		}),
	}

	return l, nil
}

func (c *Lib) SendMessage(ctx context.Context, to, msg string) error {
	p := twilioAPI.CreateMessageParams{
		To:   &to,
		From: &c.cfg.From,
		Body: &msg,
	}

	_, err := c.client.Api.CreateMessage(&p)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}
