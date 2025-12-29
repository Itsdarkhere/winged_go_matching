package twilio

import (
	"fmt"
	"math"
	"math/rand"
	"wingedapp/pgtester/internal/util/validationlib"
)

// Config contains the configuration for the Twilio client.
type Config struct {
	AccountSID string `json:"account_sid" mapstructure:"account_sid" validate:"required"`
	AuthToken  string `json:"auth_token" mapstructure:"auth_token" validate:"required"`
	From       string `json:"from" mapstructure:"from" validate:"required"`
}

func (c *Config) Validate() error {
	return validationlib.Validate(c)
}

func RandomDigitCode(length int) string {
	return fmt.Sprintf("%0*d", length, rand.Intn(int(math.Pow10(length))))
}
