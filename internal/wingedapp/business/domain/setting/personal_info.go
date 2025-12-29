package setting

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"wingedapp/pgtester/internal/util/errutil"
)

// hasDateRangeUpdateIntent checks if there's intent to update the dating preference age range
func hasDateRangeUpdateIntent(u *UpdatePersonalInformation) bool {
	a := u.DatingPreferenceAgeRangeStart.Valid
	b := u.DatingPreferenceAgeRangeEnd.Valid
	return a && b
}

// hasDateRangeUpdateErrors checks for logical errors, and domain errors re updating dt ranges
func (p *Business) hasDateRangeUpdateErrors(u *UpdatePersonalInformation) error {
	rStart := u.DatingPreferenceAgeRangeStart.Int
	rEnd := u.DatingPreferenceAgeRangeEnd.Int

	if rStart > rEnd {
		return ErrDatePrefAgeEndGtStart
	}
	if rStart < p.cfg.ThresholdMinimumAge {
		return ErrBelowMinAgeRange
	}
	return nil
}

// collectUpdatePersInfoErrors collects validation errors for UpdatePersonalInformation
func (p *Business) collectUpdatePersInfoErrors(u *UpdatePersonalInformation) errutil.List {
	errs := errutil.List{}

	if u.Name.Valid && strings.TrimSpace(u.Name.String) == "" {
		errs.Add("name is empty")
	}
	if u.Birthday.Valid && u.Birthday.Time.IsZero() {
		errs.Add("birthday is empty")
	}
	if u.Number.Valid && strings.TrimSpace(u.Number.String) == "" {
		errs.Add("number is empty")
	}
	if u.Height.Valid && u.Height.Int == 0 {
		errs.Add("height is empty")
	}
	if len(u.DatingPreferences) > 0 {
		if err := validateDatingPrefs(u.DatingPreferences); err.HasErrors() {
			errs.Add(err.Error().Error())
		} else {
			u.updateDatingPrefs = true
		}
	}
	if hasDateRangeUpdateIntent(u) {
		if err := p.hasDateRangeUpdateErrors(u); err != nil {
			joinErrs := errors.Join(
				errors.New("date pref has errors"),
				err,
			)
			errs.AddErr(joinErrs)
		}
	}
	if u.Location.Valid && strings.TrimSpace(u.Location.String) == "" {
		errs.Add("location is empty")
	}

	return errs
}

func (p *Business) UpdatePersonalInformation(ctx context.Context, params *UpdatePersonalInformation) (*PersonalInformation, error) {
	errs := p.collectUpdatePersInfoErrors(params)
	if errs.HasErrors() {
		return nil, errutil.NewValidationErrorFromList("validation failed", &errs)
	}

	tx, err := p.trans.TX()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer p.trans.Rollback(tx)

	if err = p.storer.UpdatePersonalInformation(ctx, tx, params); err != nil {
		return nil, fmt.Errorf("storer update personal info: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	personalInfo, err := p.storer.PersonalInformation(ctx, p.trans.DB(), params.UserID)
	if err != nil {
		return nil, fmt.Errorf("storer personal info: %w", err)
	}

	return personalInfo, nil
}

func (p *Business) PersonalInformation(ctx context.Context, userID string) (*PersonalInformation, error) {
	personalInfo, err := p.storer.PersonalInformation(ctx, p.trans.DB(), userID)
	if err != nil {
		return nil, fmt.Errorf("storer personalInfo: %w", err)
	}
	return personalInfo, nil
}
