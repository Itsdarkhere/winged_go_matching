package repo

import (
	"context"
	"errors"
	"fmt"
	"time"
	"wingedapp/pgtester/internal/wingedapp/business/sdk"
	"wingedapp/pgtester/internal/wingedapp/db/boilhelper"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/jackc/pgx/v4"
)

type UserInviteCodeQueryFilter struct {
	ID        null.String `json:"id"`
	Code      null.String `json:"code"`
	ForNumber null.String `json:"for_number"`

	Pagination *sdk.Pagination
	OrderBy    null.String `json:"order_by"`
	Sort       null.String `json:"sort"`
}

func userInviteCodesQMods(f *UserInviteCodeQueryFilter, paginated bool) []qm.QueryMod {
	qMods := make([]qm.QueryMod, 0)
	if paginated {
		qMods = boilhelper.ApplyPagination(qMods, f.Pagination)

		if f.OrderBy.Valid {
			sortBy := boilhelper.SortByAscOrDesc(f.Sort)

			if f.OrderBy.String == pgmodel.UserInviteCodeColumns.CreatedAt {
				qMods = append(qMods, qm.OrderBy(fmt.Sprintf(
					"%s.%s %s",
					pgmodel.TableNames.UserInviteCode,
					f.OrderBy.String, sortBy),
				))
			}
		}
	}

	if f.ID.Valid {
		qMods = append(qMods, pgmodel.UserInviteCodeWhere.ID.EQ(f.ID.String))
	}
	if f.Code.Valid {
		qMods = append(qMods, pgmodel.UserInviteCodeWhere.InviteCode.EQ(f.Code.String))
	}
	if f.ForNumber.Valid {
		qMods = append(qMods, pgmodel.UserInviteCodeWhere.ForNumber.EQ(f.ForNumber))
	}

	return qMods
}

type UserInviteCode struct {
	ID                 string      `boil:"id" json:"id" toml:"id" yaml:"id"`
	InviteCode         string      `boil:"invite_code" json:"invite_code" toml:"invite_code" yaml:"invite_code"`
	UsageCount         int         `boil:"usage_count" json:"usage_count" toml:"usage_count" yaml:"usage_count"`
	CreatedAt          time.Time   `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	ReferralSource     string      `boil:"referral_source" json:"referral_source" toml:"referral_source" yaml:"referral_source"`
	InviteCodeType     string      `boil:"invite_code_type" json:"invite_code_type" toml:"invite_code_type" yaml:"invite_code_type"` // String enum
	ForNumber          null.String `boil:"for_number" json:"for_number,omitempty" toml:"for_number" yaml:"for_number,omitempty"`
	ForNumberHash      null.String `boil:"for_number_hash" json:"for_number_hash,omitempty" toml:"for_number_hash" yaml:"for_number_hash,omitempty"`
	ReferrerNumberHash null.String `boil:"referrer_number_hash" json:"referrer_number_hash,omitempty" toml:"referrer_number_hash" yaml:"referrer_number_hash,omitempty"`
	LastUsed           null.Time   `boil:"last_used" json:"last_used,omitempty" toml:"last_used" yaml:"last_used,omitempty"`
}

func (s *Store) UserInviteCodes(ctx context.Context,
	exec boil.ContextExecutor,
	f *UserInviteCodeQueryFilter,
) ([]UserInviteCode, error) {
	var res []UserInviteCode

	// No join needed - invite_code_type is now a string enum stored directly
	err := pgmodel.UserInviteCodes(append(
		userInviteCodesQMods(f, true),
		qm.Select(boilhelper.QmSelect([]boilhelper.QmColSet{
			{
				TableName: pgmodel.TableNames.UserInviteCode,
				Cols: []boilhelper.QmCol{
					{Name: pgmodel.UserInviteCodeColumns.ID},
					{Name: pgmodel.UserInviteCodeColumns.CreatedAt},
					{Name: pgmodel.UserInviteCodeColumns.InviteCode},
					{Name: pgmodel.UserInviteCodeColumns.UsageCount},
					{Name: pgmodel.UserInviteCodeColumns.ReferralSource},
					{Name: pgmodel.UserInviteCodeColumns.ReferrerNumberHash},
					{Name: pgmodel.UserInviteCodeColumns.InviteCodeType},
					{Name: pgmodel.UserInviteCodeColumns.ForNumber},
					{Name: pgmodel.UserInviteCodeColumns.ForNumberHash},
					{Name: pgmodel.UserInviteCodeColumns.LastUsed},
				},
			},
		})...),
	)...).Bind(ctx, exec, &res)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return res, nil
		}
		return nil, fmt.Errorf("query user invite codes: %w", err)
	}

	totalCount, err := pgmodel.UserInviteCodes(userInviteCodesQMods(f, false)...).Count(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("count user invite codes: %w", err)
	}
	f.Pagination.Recalculate(int(totalCount))
	return res, nil
}

// DeleteInviteCode deletes a code.
func (s *Store) DeleteInviteCode(ctx context.Context,
	exec boil.ContextExecutor,
	id string,
) (int64, error) {
	inviteCode := pgmodel.UserInviteCode{
		ID: id,
	}
	countDeleted, err := inviteCode.Delete(ctx, exec)
	if err != nil {
		return 0, fmt.Errorf("delete invite code: %w", err)
	}
	return countDeleted, nil
}

// UserInviteCode retrieves a single invite from the database.
func (s *Store) UserInviteCode(ctx context.Context,
	exec boil.ContextExecutor,
	filter *UserInviteCodeQueryFilter,
) (*UserInviteCode, error) {
	inviteCodes, err := s.UserInviteCodes(ctx, exec, filter)
	if err != nil {
		return nil, fmt.Errorf("invite codes query: %w", err)
	}

	if len(inviteCodes) == 0 {
		return &UserInviteCode{}, nil
	}

	if len(inviteCodes) != 1 {
		return nil, fmt.Errorf("invite count mismatch, have %d, want 1", len(inviteCodes))
	}

	return &inviteCodes[0], nil
}

// UpdateUserInviteCode updates a user invite code in the database.
func (s *Store) UpdateUserInviteCode(ctx context.Context, exec boil.ContextExecutor, updater *UpdateUserInviteCode) error {
	userInviteCode, err := pgmodel.FindUserInviteCode(ctx, exec, updater.ID)
	if err != nil {
		return fmt.Errorf("find user by UserID: %w", err)
	}

	if updater.UsageCount.Valid {
		userInviteCode.UsageCount = updater.UsageCount.Int
	}
	if updater.LastUsed.Valid {
		userInviteCode.LastUsed = updater.LastUsed
	}

	if _, err := userInviteCode.Update(ctx, exec, boil.Infer()); err != nil {
		return fmt.Errorf("update user invite code: %w", err)
	}

	return nil
}

func (s *Store) UpsertUserInviteCode(ctx context.Context, exec boil.ContextExecutor, upserter *UpsertUserInviteCode) error {
	userInviteCode := pgmodel.UserInviteCode{}
	if upserter.UsageCount.Valid {
		userInviteCode.UsageCount = upserter.UsageCount.Int
	}

	cols := []string{
		pgmodel.UserInviteCodeColumns.InviteCode,
	}

	if err := userInviteCode.Upsert(ctx, exec, true, cols, boil.Infer(), boil.Infer()); err != nil {
		return fmt.Errorf("upsert user: %w", err)
	}

	return nil
}

func (s *Store) InsertUserInviteCode(ctx context.Context,
	exec boil.ContextExecutor,
	inserter *InsertUserInviteCode,
) (*pgmodel.UserInviteCode, error) {
	i := &pgmodel.UserInviteCode{
		InviteCode:         inserter.InviteCode,
		ForNumber:          inserter.ForNumber,
		ForNumberHash:      inserter.ForNumberHash,
		ReferrerNumberHash: inserter.ReferrerNumberHash,
		ReferralSource:     inserter.ReferralSource,
		InviteCodeType:     inserter.InviteCodeType, // String enum
	}
	if err := i.Insert(ctx, exec, boil.Infer()); err != nil {
		return nil, fmt.Errorf("insert user invite code: %w", err)
	}

	return i, nil
}

type UpdateUserInviteCode struct {
	ID         string
	UsageCount null.Int
	LastUsed   null.Time
}

type InsertUserInviteCode struct {
	InviteCode         string
	ReferralSource     string
	InviteCodeType     string // String enum
	ForNumber          null.String
	ForNumberHash      null.String
	ReferrerNumberHash null.String
}

type UpsertUserInviteCode struct {
	UsageCount null.Int `json:"usage_count"`
}
