package store

import (
	"wingedapp/pgtester/internal/wingedapp/db/repo"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
)

type EconomyStores struct {
	MessageStore      *MessageStore
	UserTotalsStore   *UserTotalsStore
	ActionLogStore    *ActionLogStore
	SubscriptionStore *SubscriptionStore
	TransactionStore  *TransactionStore
	DailyCheckinStore *DailyCheckinStore
	InviteCodeStore   *InviteCodeStore
	UserStore         *UserStore
	ExpiryStore       *ExpiryStore
}

func NewEconomyStores(l applog.Logger) *EconomyStores {
	r := &repo.Store{}
	return &EconomyStores{
		MessageStore:      &MessageStore{l, r},
		UserTotalsStore:   &UserTotalsStore{l, r},
		SubscriptionStore: NewSubscriptionStore(l),
		ActionLogStore:    &ActionLogStore{l, r},
		TransactionStore:  &TransactionStore{l, r},
		DailyCheckinStore: NewDailyCheckinStore(l),
		InviteCodeStore:   NewInviteCodeStore(l, r),
		UserStore:         NewUserStore(l, r),
		ExpiryStore:       NewExpiryStore(),
	}
}
