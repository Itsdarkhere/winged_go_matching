package repo

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/ericlagergren/decimal"
)

// InsertVenue represents a venue to be inserted.
type InsertVenue struct {
	ExternalProvider string
	ExternalID       string
	Name             string
	DisplayName      null.String
	Address          null.String
	Latitude         null.Float64
	Longitude        null.Float64
	VenueData        null.JSON
	DietaryTags      []string
	DateTypeFit      []string
	RefreshAfter     null.Time
}

// InsertVenue inserts a new venue.
func (s *Store) InsertVenue(
	ctx context.Context,
	exec boil.ContextExecutor,
	inserter *InsertVenue,
) (*pgmodel.Venue, error) {
	if inserter.ExternalProvider == "" {
		return nil, fmt.Errorf("external_provider is required")
	}
	if inserter.ExternalID == "" {
		return nil, fmt.Errorf("external_id is required")
	}
	if inserter.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	v := &pgmodel.Venue{
		ExternalProvider: inserter.ExternalProvider,
		ExternalID:       inserter.ExternalID,
		Name:             inserter.Name,
		DisplayName:      inserter.DisplayName,
		Address:          inserter.Address,
		VenueData:        inserter.VenueData,
		RefreshAfter:     inserter.RefreshAfter,
	}

	if inserter.Latitude.Valid {
		v.Latitude = types.NewNullDecimal(new(decimal.Big).SetFloat64(inserter.Latitude.Float64))
	}
	if inserter.Longitude.Valid {
		v.Longitude = types.NewNullDecimal(new(decimal.Big).SetFloat64(inserter.Longitude.Float64))
	}
	if len(inserter.DietaryTags) > 0 {
		v.DietaryTags = inserter.DietaryTags
	}
	if len(inserter.DateTypeFit) > 0 {
		v.DateTypeFit = inserter.DateTypeFit
	}

	if err := v.Insert(ctx, exec, boil.Infer()); err != nil {
		return nil, fmt.Errorf("insert venue: %w", err)
	}

	return v, nil
}

// UpdateVenue contains optional fields for updating a venue.
type UpdateVenue struct {
	ID           string
	DisplayName  null.String
	Address      null.String
	Latitude     null.Float64
	Longitude    null.Float64
	VenueData    null.JSON
	DietaryTags  []string
	DateTypeFit  []string
	RefreshAfter null.Time
}

// UpdateVenue updates an existing venue.
func (s *Store) UpdateVenue(
	ctx context.Context,
	exec boil.ContextExecutor,
	updater *UpdateVenue,
) (*pgmodel.Venue, error) {
	if updater.ID == "" {
		return nil, fmt.Errorf("id is required")
	}

	v, err := pgmodel.FindVenue(ctx, exec, updater.ID)
	if err != nil {
		return nil, fmt.Errorf("find venue: %w", err)
	}

	cols := make([]string, 0)

	if updater.DisplayName.Valid {
		v.DisplayName = updater.DisplayName
		cols = append(cols, pgmodel.VenueColumns.DisplayName)
	}
	if updater.Address.Valid {
		v.Address = updater.Address
		cols = append(cols, pgmodel.VenueColumns.Address)
	}
	if updater.Latitude.Valid {
		v.Latitude = types.NewNullDecimal(new(decimal.Big).SetFloat64(updater.Latitude.Float64))
		cols = append(cols, pgmodel.VenueColumns.Latitude)
	}
	if updater.Longitude.Valid {
		v.Longitude = types.NewNullDecimal(new(decimal.Big).SetFloat64(updater.Longitude.Float64))
		cols = append(cols, pgmodel.VenueColumns.Longitude)
	}
	if updater.VenueData.Valid {
		v.VenueData = updater.VenueData
		cols = append(cols, pgmodel.VenueColumns.VenueData)
	}
	if len(updater.DietaryTags) > 0 {
		v.DietaryTags = updater.DietaryTags
		cols = append(cols, pgmodel.VenueColumns.DietaryTags)
	}
	if len(updater.DateTypeFit) > 0 {
		v.DateTypeFit = updater.DateTypeFit
		cols = append(cols, pgmodel.VenueColumns.DateTypeFit)
	}
	if updater.RefreshAfter.Valid {
		v.RefreshAfter = updater.RefreshAfter
		cols = append(cols, pgmodel.VenueColumns.RefreshAfter)
	}

	if len(cols) == 0 {
		return v, nil
	}

	if _, err := v.Update(ctx, exec, boil.Whitelist(cols...)); err != nil {
		return nil, fmt.Errorf("update venue: %w", err)
	}

	return v, nil
}
