package store

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/db/repo"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
	"wingedapp/pgtester/internal/wingedapp/lib/economy"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

type SubscriptionStore struct {
	logger applog.Logger
	repo   *repo.Store
}

func NewSubscriptionStore(l applog.Logger) *SubscriptionStore {
	return &SubscriptionStore{
		logger: l,
		repo:   &repo.Store{},
	}
}

func (s *SubscriptionStore) SubscriptionPlan(
	ctx context.Context,
	exec boil.ContextExecutor,
	subscriptionType,
	subscriptionName string,
) (*economy.SubscriptionPlan, error) {
	var subPlan economy.SubscriptionPlan

	// tables
	wingsSubPlanTbl := pgmodel.TableNames.WingsEcnSubscriptionPlan

	// cols
	wingsSubCols := pgmodel.WingsEcnSubscriptionPlanColumns

	id := wingsSubCols.ID
	name := wingsSubCols.Name
	price := wingsSubCols.Price
	wings := wingsSubCols.Wings
	subType := wingsSubCols.SubscriptionType

	if err := pgmodel.NewQuery(
		qm.Select(
			"wsp."+id+" AS "+id,
			"wsp."+name+" AS "+name,
			"wsp."+price+" AS "+price,
			"wsp."+wings+" AS "+wings,
		),
		qm.From(wingsSubPlanTbl+" wsp"),
		qm.Where("wsp."+subType+" = ?", subscriptionType),
		qm.Where("wsp."+name+" = ?", subscriptionName),
	).Bind(ctx, exec, &subPlan); err != nil {
		return nil, fmt.Errorf("find subscription plan: %w", err)
	}

	return &subPlan, nil
}
