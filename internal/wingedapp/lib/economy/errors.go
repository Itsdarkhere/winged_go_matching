package economy

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidEntryType      = errors.New("invalid entry type")
	ErrMissingJSONDetails    = errors.New("missing required JSON details")
	ErrMissingSubscriptionID = errors.New("missing subscription ID")
	ErrInsufficientWings     = errors.New("insufficient wings balance")
	ErrWeeklyCapReached      = errors.New("weekly earning cap reached")
	ErrBalanceCapReached     = errors.New("balance cap reached")
	ErrReferralCapReached    = errors.New("referral monthly cap reached")
	ErrAlreadyCheckedInToday = errors.New("already checked in today")
	ErrAlreadyProcessed      = errors.New("payment already processed")
	ErrUnknownProductID      = errors.New("unknown product ID")
)

// errInvalidAction formats an error for an invalid action type.
func errInvalidAction(entry ActionType) error {
	return fmt.Errorf("invalid entry: %s", string(entry))
}
