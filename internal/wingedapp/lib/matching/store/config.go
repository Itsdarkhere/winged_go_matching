package store

import (
	"context"
	"fmt"
	"strconv"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/db/repo"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
	"wingedapp/pgtester/internal/wingedapp/lib/matching"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/ericlagergren/decimal"
)

type ConfigStore struct {
	l    applog.Logger
	repo *repo.Store
}

func (s *ConfigStore) Configs(
	ctx context.Context,
	exec boil.ContextExecutor,
	f *matching.QueryFilterMatchConfig,
) ([]matching.Config, error) {
	var configs []matching.Config

	c := pgmodel.MatchConfigColumns

	qMods := append(
		qModsMatchConfig(f),
		qm.Select(
			"mc."+c.ID+" AS id",
			"COALESCE(mc."+c.AgeRangeStart+", 0) AS age_range_start",
			"COALESCE(mc."+c.AgeRangeEnd+", 0) AS age_range_end",
			"mc."+c.AgeRangeWomanOlderBy+" AS age_range_woman_older_by",
			"mc."+c.AgeRangeManOlderBy+" AS age_range_man_older_by",
			"mc."+c.HeightMaleGreaterByCM+"::float8 AS height_male_greater_by_cm",
			"mc."+c.LocationRadiusKM+"::float8 AS location_radius_km",
			"mc."+c.LocationAdaptiveExpansion+" AS location_adaptive_expansion",
			"mc."+c.DropHours+" AS drop_hours",
			"mc."+c.DropHoursUtc+" AS drop_hours_utc",
			"mc."+c.StaleChatAgentSetup+" AS stale_chat_agent_setup",
			"mc."+c.StaleChatNudge+" AS stale_chat_nudge",
			"mc."+c.MatchExpirationHours+" AS match_expiration_hours",
			"mc."+c.MatchBlockDeclined+" AS match_block_declined",
			"mc."+c.MatchBlockIgnored+" AS match_block_ignored",
			"mc."+c.MatchBlockClosed+" AS match_block_closed",
			"mc."+c.ScoreRangeStart+"::float8 AS score_range_start",
			"mc."+c.ScoreRangeEnd+"::float8 AS score_range_end",
		),
		qm.From(pgmodel.TableNames.MatchConfig+" mc"),
	)

	if err := pgmodel.NewQuery(qMods...).Bind(ctx, exec, &configs); err != nil {
		return nil, fmt.Errorf("configs: %w", err)
	}

	return configs, nil
}

func (s *ConfigStore) Config(
	ctx context.Context,
	exec boil.ContextExecutor,
	f *matching.QueryFilterMatchConfig,
) (*matching.Config, error) {
	configs, err := s.Configs(ctx, exec, f)
	if err != nil {
		return nil, fmt.Errorf("query configs: %w", err)
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("no config found")
	}

	if len(configs) > 1 {
		return nil, fmt.Errorf("config count mismatch, have %d, want 1", len(configs))
	}

	return &configs[0], nil
}

func qModsMatchConfig(f *matching.QueryFilterMatchConfig) []qm.QueryMod {
	qMods := make([]qm.QueryMod, 0)

	c := pgmodel.MatchConfigColumns

	if f == nil {
		return qMods
	}

	// ID filter
	if f.ID.Valid {
		qMods = append(qMods, qm.Where("mc."+c.ID+" = ?", f.ID.String))
	}

	// Age range filters
	if f.AgeRangeStartMin.Valid {
		qMods = append(qMods, qm.Where("COALESCE(mc."+c.AgeRangeStart+", 0) >= ?", f.AgeRangeStartMin.Int))
	}
	if f.AgeRangeStartMax.Valid {
		qMods = append(qMods, qm.Where("COALESCE(mc."+c.AgeRangeStart+", 0) <= ?", f.AgeRangeStartMax.Int))
	}
	if f.AgeRangeEndMin.Valid {
		qMods = append(qMods, qm.Where("COALESCE(mc."+c.AgeRangeEnd+", 0) >= ?", f.AgeRangeEndMin.Int))
	}
	if f.AgeRangeEndMax.Valid {
		qMods = append(qMods, qm.Where("COALESCE(mc."+c.AgeRangeEnd+", 0) <= ?", f.AgeRangeEndMax.Int))
	}

	// Height filter
	if f.HeightMaleGreaterByCMMin.Valid {
		qMods = append(qMods, qm.Where("mc."+c.HeightMaleGreaterByCM+" >= ?", f.HeightMaleGreaterByCMMin.Float64))
	}
	if f.HeightMaleGreaterByCMMax.Valid {
		qMods = append(qMods, qm.Where("mc."+c.HeightMaleGreaterByCM+" <= ?", f.HeightMaleGreaterByCMMax.Float64))
	}

	// Location radius filter
	if f.LocationRadiusKMMin.Valid {
		qMods = append(qMods, qm.Where("mc."+c.LocationRadiusKM+" >= ?", f.LocationRadiusKMMin.Float64))
	}
	if f.LocationRadiusKMMax.Valid {
		qMods = append(qMods, qm.Where("mc."+c.LocationRadiusKM+" <= ?", f.LocationRadiusKMMax.Float64))
	}

	// Drop hours filters
	if f.DropHour.Valid {
		qMods = append(qMods, qm.Where("? = ANY(mc."+c.DropHours+")", f.DropHour.String))
	}
	if f.DropHourUTC.Valid {
		qMods = append(qMods, qm.Where("? = ANY(mc."+c.DropHoursUtc+")", f.DropHourUTC.String))
	}
	if f.HasDropHours.Valid {
		if f.HasDropHours.Bool {
			qMods = append(qMods, qm.Where("array_length(mc."+c.DropHours+", 1) > 0"))
		} else {
			qMods = append(qMods, qm.Where("(mc."+c.DropHours+" IS NULL OR array_length(mc."+c.DropHours+", 1) IS NULL)"))
		}
	}

	// Stale chat filters
	if f.StaleChatNudgeMin.Valid {
		qMods = append(qMods, qm.Where("mc."+c.StaleChatNudge+" >= ?", f.StaleChatNudgeMin.Int))
	}
	if f.StaleChatNudgeMax.Valid {
		qMods = append(qMods, qm.Where("mc."+c.StaleChatNudge+" <= ?", f.StaleChatNudgeMax.Int))
	}
	if f.StaleChatAgentSetupMin.Valid {
		qMods = append(qMods, qm.Where("mc."+c.StaleChatAgentSetup+" >= ?", f.StaleChatAgentSetupMin.Int))
	}
	if f.StaleChatAgentSetupMax.Valid {
		qMods = append(qMods, qm.Where("mc."+c.StaleChatAgentSetup+" <= ?", f.StaleChatAgentSetupMax.Int))
	}

	// Match expiration filter
	if f.MatchExpirationHoursMin.Valid {
		qMods = append(qMods, qm.Where("mc."+c.MatchExpirationHours+" >= ?", f.MatchExpirationHoursMin.Int))
	}
	if f.MatchExpirationHoursMax.Valid {
		qMods = append(qMods, qm.Where("mc."+c.MatchExpirationHours+" <= ?", f.MatchExpirationHoursMax.Int))
	}

	// Score range filters
	if f.ScoreRangeStartMin.Valid {
		qMods = append(qMods, qm.Where("mc."+c.ScoreRangeStart+" >= ?", f.ScoreRangeStartMin.Float64))
	}
	if f.ScoreRangeStartMax.Valid {
		qMods = append(qMods, qm.Where("mc."+c.ScoreRangeStart+" <= ?", f.ScoreRangeStartMax.Float64))
	}
	if f.ScoreRangeEndMin.Valid {
		qMods = append(qMods, qm.Where("mc."+c.ScoreRangeEnd+" >= ?", f.ScoreRangeEndMin.Float64))
	}
	if f.ScoreRangeEndMax.Valid {
		qMods = append(qMods, qm.Where("mc."+c.ScoreRangeEnd+" <= ?", f.ScoreRangeEndMax.Float64))
	}

	// Sorting with column whitelisting
	if f.OrderBy.Valid && f.Sort.Valid {
		allowedColumns := map[string]string{
			"age_range_start":        c.AgeRangeStart,
			"age_range_end":          c.AgeRangeEnd,
			"height_male_greater_by": c.HeightMaleGreaterByCM,
			"location_radius_km":     c.LocationRadiusKM,
			"stale_chat_nudge":       c.StaleChatNudge,
			"match_expiration_hours": c.MatchExpirationHours,
			"score_range_start":      c.ScoreRangeStart,
			"score_range_end":        c.ScoreRangeEnd,
		}

		if col, ok := allowedColumns[f.OrderBy.String]; ok {
			orderDir := "DESC"
			if f.Sort.String == "+" {
				orderDir = "ASC"
			}
			qMods = append(qMods, qm.OrderBy("mc."+col+" "+orderDir))
		}
	}

	return qMods
}

// Update updates the match configuration with the provided fields.
// Only non-nil/valid fields are updated (PATCH semantics).
// Returns the updated configuration.
func (s *ConfigStore) Update(
	ctx context.Context,
	exec boil.ContextExecutor,
	updater *matching.UpdateMatchConfig,
) (*matching.Config, error) {
	// Fetch the singleton config (match_config typically has one row)
	existing, err := pgmodel.MatchConfigs().One(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", matching.ErrConfigNotFound, err)
	}

	// Apply updates only for provided fields
	if updater.AgeRangeStart.Valid {
		existing.AgeRangeStart.SetValid(updater.AgeRangeStart.Int)
	}
	if updater.AgeRangeEnd.Valid {
		existing.AgeRangeEnd.SetValid(updater.AgeRangeEnd.Int)
	}
	if updater.AgeRangeWomanOlderBy.Valid {
		existing.AgeRangeWomanOlderBy = updater.AgeRangeWomanOlderBy.Int
	}
	if updater.AgeRangeManOlderBy.Valid {
		existing.AgeRangeManOlderBy = updater.AgeRangeManOlderBy.Int
	}
	if updater.HeightMaleGreaterByCM.Valid {
		existing.HeightMaleGreaterByCM = decimalFromFloat64(updater.HeightMaleGreaterByCM.Float64)
	}
	if updater.LocationRadiusKM.Valid {
		existing.LocationRadiusKM = decimalFromFloat64(updater.LocationRadiusKM.Float64)
	}
	if updater.LocationAdaptiveExpansion != nil {
		existing.LocationAdaptiveExpansion = *updater.LocationAdaptiveExpansion
	}
	if updater.DropHours != nil {
		existing.DropHours = *updater.DropHours
	}
	if updater.DropHoursUTC != nil {
		existing.DropHoursUtc = *updater.DropHoursUTC
	}
	if updater.StaleChatNudge.Valid {
		existing.StaleChatNudge = updater.StaleChatNudge.Int
	}
	if updater.StaleChatAgentSetup.Valid {
		existing.StaleChatAgentSetup = updater.StaleChatAgentSetup.Int
	}
	if updater.MatchExpirationHours.Valid {
		existing.MatchExpirationHours = updater.MatchExpirationHours.Int
	}
	if updater.MatchBlockDeclined.Valid {
		existing.MatchBlockDeclined = updater.MatchBlockDeclined.Int
	}
	if updater.MatchBlockIgnored.Valid {
		existing.MatchBlockIgnored = updater.MatchBlockIgnored.Int
	}
	if updater.MatchBlockClosed.Valid {
		existing.MatchBlockClosed = updater.MatchBlockClosed.Int
	}
	if updater.ScoreRangeStart.Valid {
		existing.ScoreRangeStart = decimalFromFloat64(updater.ScoreRangeStart.Float64)
	}
	if updater.ScoreRangeEnd.Valid {
		existing.ScoreRangeEnd = decimalFromFloat64(updater.ScoreRangeEnd.Float64)
	}

	// Persist the updated config
	_, err = existing.Update(ctx, exec, boil.Infer())
	if err != nil {
		return nil, fmt.Errorf("update config: %w", err)
	}

	// Fetch and return the updated config using our domain model
	return s.Config(ctx, exec, nil)
}

func decimalFromFloat64(f float64) types.Decimal {
	d := new(decimal.Big)
	d, _ = d.SetString(strconv.FormatFloat(f, 'f', -1, 64))
	return types.NewDecimal(d)
}
