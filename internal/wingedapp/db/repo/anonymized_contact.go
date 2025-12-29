package repo

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/business/sdk"
	"wingedapp/pgtester/internal/wingedapp/db/boilhelper"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

type QueryFilterAnonymizedContact struct {
	ContactHashes []string

	Pagination *sdk.Pagination
	OrderBy    null.String `json:"order_by"`
	Sort       null.String `json:"sort"`
}

// qModsAnonymizedContacts is a helper to build query mods for UserAIConvo queries
func qModsAnonymizedContacts(f *QueryFilterAnonymizedContact, paginated bool) []qm.QueryMod {
	qMods := make([]qm.QueryMod, 0)
	if paginated {
		qMods = boilhelper.ApplyPagination(qMods, f.Pagination)
	}

	if f.OrderBy.Valid {
		// apply sort to all order bys
		sort := "DESC"
		if f.Sort.Valid && f.Sort.String == "+" {
			sort = "ASC"
		}

		if f.OrderBy.String == pgmodel.AnonymizedContactColumns.CreatedAt {
			qMods = append(qMods, qm.OrderBy(fmt.Sprintf(
				"%s.%s %s",
				pgmodel.TableNames.AnonymizedContact,
				f.OrderBy.String, sort),
			))
		}
	}

	if len(f.ContactHashes) > 0 {
		qMods = append(qMods, pgmodel.AnonymizedContactWhere.ContactHash.IN(f.ContactHashes))
	}

	return qMods
}

type AnonymizedContact struct {
	ContactHash string `json:"hash" boil:"contact_hash"`
	Count       int    `json:"count" boil:"count"`
	IsMember    bool   `json:"is_member" boil:"is_member"`
	IsInvited   bool   `json:"is_invited" boil:"is_invited"`
}

// AnonymizedContactsByCount retrieves anonymized
// contacts grouped by ContactHash counts.
// Too complex — probably won't repeat this approach.
// Need better reference code.
func (s *Store) AnonymizedContactsByCount(ctx context.Context,
	exec boil.ContextExecutor,
	f *QueryFilterAnonymizedContact,
) ([]AnonymizedContact, error) {
	var bind []AnonymizedContact

	// query intentionally aggregates "invited" and "is member" status
	// to make it simple.
	// Status translation is done in business's store layer to make this purely fetcher.
	qMods := func() []qm.QueryMod {
		qModBase := qModsAnonymizedContacts(f, true)
		qModCols := qm.Select(
			boilhelper.QmSelect([]boilhelper.QmColSet{
				{
					TableName: pgmodel.TableNames.AnonymizedContact,
					Cols: []boilhelper.QmCol{ // local cols
						{
							Name:     pgmodel.AnonymizedContactColumns.ID,
							Modifier: "COUNT(DISTINCT(%s))",
							Alias:    "count",
						},
						{Name: pgmodel.AnonymizedContactColumns.ContactHash},
					},
				},
				// joined I guess
				{
					TableName: pgmodel.TableNames.Users,
					Cols: []boilhelper.QmCol{ // local cols
						{
							Name:     pgmodel.UserColumns.Sha256Hash,
							Modifier: "COUNT(DISTINCT(%s)) > 0",
							Alias:    "is_member",
						},
					},
				},
				{
					TableName: pgmodel.TableNames.UserInviteCode,
					Cols: []boilhelper.QmCol{ // local cols
						{
							Name:     pgmodel.UserInviteCodeColumns.ReferralSource,
							Modifier: "COUNT(DISTINCT(%s)) > 0",
							Alias:    "is_invited",
						},
					},
				},
			},
			)...,
		)
		qModGroupBy := boilhelper.QmGroupBy(
			pgmodel.TableNames.AnonymizedContact,
			pgmodel.AnonymizedContactColumns.ContactHash,
		)

		// left join users — identify "members"
		qModLJoinUsers := qm.LeftOuterJoin(boilhelper.QmJoinStr(
			pgmodel.TableNames.Users,
			pgmodel.UserColumns.Sha256Hash,
			pgmodel.TableNames.AnonymizedContact,
			pgmodel.AnonymizedContactColumns.ContactHash),
		)

		// left join user invite codes — identify if "invited"
		qModLJoinUserInviteCode := qm.LeftOuterJoin(boilhelper.QmJoinStr(
			pgmodel.TableNames.UserInviteCode,
			pgmodel.UserInviteCodeColumns.ForNumberHash,
			pgmodel.TableNames.AnonymizedContact,
			pgmodel.AnonymizedContactColumns.ContactHash),
		)

		return append(
			qModBase,
			qModCols,
			qModGroupBy,
			qModLJoinUsers,
			qModLJoinUserInviteCode,
		)
	}

	qModCount := func() []qm.QueryMod {
		qModBase := qModsAnonymizedContacts(f, true)
		qModCols := qm.Select(
			boilhelper.QmSelect([]boilhelper.QmColSet{
				{
					TableName: pgmodel.TableNames.AnonymizedContact,
					Cols: []boilhelper.QmCol{ // local cols
						{
							Name:     pgmodel.AnonymizedContactColumns.ID,
							Modifier: "DISTINCT(%s)",
							Alias:    "id",
						},
						{Name: pgmodel.AnonymizedContactColumns.ContactHash},
					},
				},
			},
			)...,
		)

		return append(
			qModBase,
			qModCols,
		)
	}

	if err := pgmodel.AnonymizedContacts(qMods()...).Bind(ctx, exec, &bind); err != nil {
		return nil, fmt.Errorf("query anonymized contacts: %w", err)
	}

	count, err := pgmodel.AnonymizedContacts(qModCount()...).Count(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("count anonymized countacts: %w", err)
	}
	f.Pagination.Recalculate(int(count))

	return bind, nil
}

type InsertAnonymizedContact struct {
	OwnerHash   string
	ContactHash string
}

func (s *Store) UpsertAnonymizedContact(ctx context.Context,
	db boil.ContextExecutor,
	insert *InsertAnonymizedContact,
) (*pgmodel.AnonymizedContact, error) {
	ac := &pgmodel.AnonymizedContact{
		OwnerHash:   insert.OwnerHash,
		ContactHash: insert.ContactHash,
	}

	conflictCols := []string{
		pgmodel.AnonymizedContactColumns.ContactHash,
		pgmodel.AnonymizedContactColumns.OwnerHash,
	}
	if err := ac.Upsert(ctx, db, true, conflictCols, boil.Infer(), boil.Infer()); err != nil {
		return nil, fmt.Errorf("insert anon contact failed: %w", err)
	}
	return ac, nil
}
