package validationlib

import (
	"regexp"

	"github.com/ttacon/libphonenumber"
)

var digit6Regex = regexp.MustCompile(`^\d{6}$`)

// sha256Regex is a regex pattern to validate sha256 strings.
var sha256Regex = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)

// isValidE164 checks if a phone number is valid according to the E.164 format.
func isValidE164(number string) bool {
	num, err := libphonenumber.Parse(number, "")
	if err != nil {
		return false
	}
	return libphonenumber.IsValidNumber(num)
}
