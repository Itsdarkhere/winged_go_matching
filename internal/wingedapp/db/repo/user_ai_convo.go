package repo

import (
	"context"
	"errors"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/business/sdk"
	"wingedapp/pgtester/internal/wingedapp/db/boilhelper"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/jackc/pgx/v4"
)

// qModsUserAIConvo is a helper to build query mods for UserAIConvo queries
func qModsUserAIConvo(f *QueryFilterUserAIConvo, paginated bool) []qm.QueryMod {
	qMods := make([]qm.QueryMod, 0)
	if paginated {
		qMods = boilhelper.ApplyPagination(qMods, f.Pagination)

		if f.OrderBy.Valid {
			// apply sort to all order bys
			sort := "DESC"
			if f.Sort.Valid && f.Sort.String == "+" {
				sort = "ASC"
			}

			allowedOrderByColumns := map[string]bool{
				pgmodel.UserAIConvoColumns.CreatedAt: true,
				pgmodel.UserAIConvoColumns.Message:   true,
			}
			if allowedOrderByColumns[f.OrderBy.String] {
				qMods = append(qMods, qm.OrderBy(fmt.Sprintf(
					"%s.%s %s",
					pgmodel.TableNames.UserAiConvo,
					f.OrderBy.String, sort),
				))
			}
		}
	}

	if f.ID.Valid {
		qMods = append(qMods, pgmodel.UserAIConvoWhere.ID.EQ(f.ID.String))
	}
	if f.UserID.Valid {
		qMods = append(qMods, pgmodel.UserAIConvoWhere.UserID.EQ(f.UserID.String))
	}

	return qMods
}

type QueryFilterUserAIConvo struct {
	ID               null.String `json:"id"`
	UserID           null.String `json:"user_id"`
	PromptResponseID null.String `json:"prompt_response_id"`

	Pagination *sdk.Pagination `json:"pagination"`
	OrderBy    null.String     `json:"order_by"`
	Sort       null.String     `json:"sort"`
}

type UserAIConvo struct {
	ID                string    `boil:"id" json:"id" toml:"id" yaml:"id"`
	UserID            string    `boil:"user_id" json:"user_id" toml:"user_id" yaml:"user_id"`
	AiConvoType       string    `boil:"ai_convo_type" json:"ai_convo_type" toml:"ai_convo_type" yaml:"ai_convo_type"` // String enum
	Message           string    `boil:"message" json:"message" toml:"message" yaml:"message"`
	AdditionalContext string    `boil:"additional_context" json:"additional_context" toml:"additional_context" yaml:"additional_context"`
	Response          string    `boil:"response" json:"response" toml:"response" yaml:"response"`
	CreatedAt         null.Time `boil:"created_at" json:"created_at,omitempty" toml:"created_at" yaml:"created_at,omitempty"`
	UpdatedAt         null.Time `boil:"updated_at" json:"updated_at,omitempty" toml:"updated_at" yaml:"updated_at,omitempty"`
}

func (s *Store) UserAIConvos(ctx context.Context, exec boil.ContextExecutor, f *QueryFilterUserAIConvo) ([]UserAIConvo, error) {
	var res []UserAIConvo

	// No join needed - ai_convo_type is now a string enum stored directly
	err := pgmodel.UserAIConvos(append(
		qModsUserAIConvo(f, true),
		qm.Select(boilhelper.QmSelect([]boilhelper.QmColSet{
			{
				TableName: pgmodel.TableNames.UserAiConvo,
				Cols: []boilhelper.QmCol{
					{Name: pgmodel.UserAIConvoColumns.ID},
					{Name: pgmodel.UserAIConvoColumns.Message},
					{Name: pgmodel.UserAIConvoColumns.AdditionalContext},
					{Name: pgmodel.UserAIConvoColumns.Response},
					{Name: pgmodel.UserAIConvoColumns.UserID},
					{Name: pgmodel.UserAIConvoColumns.AiConvoType},
					{Name: pgmodel.UserAIConvoColumns.CreatedAt},
					{Name: pgmodel.UserAIConvoColumns.UpdatedAt},
				},
			},
		})...),
	)...).Bind(ctx, exec, &res)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return res, nil
		}
		return nil, fmt.Errorf("query user ai convos: %w", err)
	}

	totalCount, err := pgmodel.UserAIConvos(qModsUserAIConvo(f, false)...).Count(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("count user ai convos: %w", err)
	}
	f.Pagination.Recalculate(int(totalCount))
	return res, nil
}

// UserAIConvo returns one user AI convo based on the filter
func (s *Store) UserAIConvo(ctx context.Context,
	exec boil.ContextExecutor,
	filter *QueryFilterUserAIConvo,
) (*UserAIConvo, error) {
	userAIConvo, err := s.UserAIConvos(ctx, exec, filter)
	if err != nil {
		return nil, fmt.Errorf("user ai convo: %w", err)
	}

	if len(userAIConvo) == 0 {
		return nil, fmt.Errorf("user ai convo: none found")
	}
	if len(userAIConvo) != 1 {
		return nil, fmt.Errorf("user ai convo count mismatch, have %d, want 1", len(userAIConvo))
	}

	return &userAIConvo[0], nil
}

func (s *Store) InsertUserAIConvo(ctx context.Context, db boil.ContextExecutor, inserter *InsertUserAIConvo) (*UserAIConvo, error) {
	if inserter.UserID == "" {
		return nil, fmt.Errorf("user UserID is required for inserting UserAIConvo")
	}

	inserted := &pgmodel.UserAIConvo{
		UserID:            inserter.UserID,
		PromptResponseID:  inserter.PromptResponseID,
		AiConvoType:       inserter.AiConvoType, // String enum
		Message:           inserter.Message,
		AdditionalContext: inserter.AdditionalContext,
		Response:          inserter.Response,
	}
	if err := inserted.Insert(ctx, db, boil.Infer()); err != nil {
		return nil, fmt.Errorf("insert UserAIConvo: %w", err)
	}

	enrichedInserted, err := s.UserAIConvo(ctx, db, &QueryFilterUserAIConvo{
		ID: null.StringFrom(inserted.ID),
	})
	if err != nil {
		return nil, fmt.Errorf("fetching inserted user ai convo: %w", err)
	}

	return enrichedInserted, nil
}

type InsertUserAIConvo struct {
	UserID            string `json:"user_id"`
	PromptResponseID  string `json:"prompt_response_id"`
	AiConvoType       string `json:"ai_convo_type"` // String enum
	Message           string `json:"message"`
	Response          string `json:"response"`
	AdditionalContext string `json:"additional_context"`
}
