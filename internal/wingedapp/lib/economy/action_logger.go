package economy

import (
	"errors"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
)

/*
	Economy module handles all logic related to earning, and spending wings.

	Coding paradigm: entrypoint pattern.
	- we register events, and trigger side effects e.g:
		- sending message flow:
			-> register at message table
			-> check the counter if its time to add wings
				-> if yes, add wings to user
*/

// ActionLogger this is called action logger because earning, and spending
// are "actions" that indirectly lead to changes in a user's wing totals.
type ActionLogger struct {
	logger applog.Logger

	/* accessors to trans table */
	userTotalsStorer userTotalsStorer
	actionLogStorer  actionLogStorer

	/* domain accessors */
	userStorer         userStorer
	messageStorer      messageStorer
	settingGetter      settingGetter
	subscriptionStorer subscriptionPlanStorer
	transactionStorer  transactionStorer
	inviteCodeStorer   inviteCodeStorer
}

func NewActionLogger(
	logger applog.Logger,
	settingGetter settingGetter,
	messageStorer messageStorer,
	userTotalsStorer userTotalsStorer,
	actionStorer actionLogStorer,
	subscriptionStorer subscriptionPlanStorer,
	transStorer transactionStorer,
	inviteCodeStorer inviteCodeStorer,
	userStorer userStorer,
) (*ActionLogger, error) {
	if logger == nil {
		return nil, errors.New("logger is required")
	}
	if messageStorer == nil {
		return nil, errors.New("messageStorer is required")
	}
	if userTotalsStorer == nil {
		return nil, errors.New("userTotalsStorer is required")
	}
	if actionStorer == nil {
		return nil, errors.New("actionLogStorer is required")
	}
	if settingGetter == nil {
		return nil, errors.New("settingGetter is required")
	}
	if subscriptionStorer == nil {
		return nil, errors.New("subscriptionPlanStorer is required")
	}
	if transStorer == nil {
		return nil, errors.New("transactionStorer is required")
	}
	if inviteCodeStorer == nil {
		return nil, errors.New("inviteCodeStorer is required")
	}
	if userStorer == nil {
		return nil, errors.New("userStorer is required")
	}

	return &ActionLogger{
		logger:             logger,
		messageStorer:      messageStorer,
		userTotalsStorer:   userTotalsStorer,
		actionLogStorer:    actionStorer,
		settingGetter:      settingGetter,
		subscriptionStorer: subscriptionStorer,
		transactionStorer:  transStorer,
		inviteCodeStorer:   inviteCodeStorer,
		userStorer:         userStorer,
	}, nil
}
