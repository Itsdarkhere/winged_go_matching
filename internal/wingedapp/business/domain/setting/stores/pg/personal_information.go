package pg

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/business/domain/setting"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/db/repo"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
)

// PersonalInformation retrieves the personal information of a user.
func (s *Store) PersonalInformation(ctx context.Context,
	exec boil.ContextExecutor,
	userID string,
) (*setting.PersonalInformation, error) {
	pgUser, err := s.repoBackendApp.User(ctx, exec, &repo.QueryFilterUser{
		ID: null.StringFrom(userID),
	})
	if err != nil {
		return nil, fmt.Errorf("user repo: %w", err)
	}

	pgUserDatingPrefs, err := s.repoBackendApp.UserDatingPreferences(ctx, exec, &repo.UserDatingPreferencesQueryFilter{
		UserID: null.StringFrom(userID),
	})
	if err != nil {
		return nil, fmt.Errorf("user dating pref repo: %w", err)
	}

	dietaryRestrictions, err := s.repoBackendApp.UserDietaryRestrictions(ctx, exec, userID)
	if err != nil {
		return nil, fmt.Errorf("user dietary restrictions repo: %w", err)
	}

	dateTypePreferences, err := s.repoBackendApp.UserDateTypePreferences(ctx, exec, userID)
	if err != nil {
		return nil, fmt.Errorf("user date type preferences repo: %w", err)
	}

	mobilityConstraints, err := s.repoBackendApp.UserMobilityConstraints(ctx, exec, userID)
	if err != nil {
		return nil, fmt.Errorf("user mobility constraints repo: %w", err)
	}

	personalInfo := storeModelsToPersonalInfo(
		pgUser,
		pgUserDatingPrefs,
		dietaryRestrictions,
		dateTypePreferences,
		mobilityConstraints,
	)

	return personalInfo, nil
}

func storeModelsToPersonalInfo(
	user *pgmodel.User,
	userDatingPrefs pgmodel.UserDatingPreferenceSlice,
	dietaryRestrictions []string,
	dateTypePreferences []string,
	mobilityConstraints []string,
) *setting.PersonalInformation {
	return &setting.PersonalInformation{
		ID:                            user.ID,
		Email:                         user.Email,
		Gender:                        user.Gender.String,
		Name:                          user.FirstName.String,
		Birthday:                      user.Birthday.Time,
		Number:                        user.MobileNumber.String,
		Height:                        user.HeightCM.Int,
		AgentDating:                   user.AgentDating.Bool,
		DatingPreferenceAgeRangeStart: user.DatingPrefAgeRangeStart.Int,
		DatingPreferenceAgeRangeEnd:   user.DatingPrefAgeRangeEnd.Int,
		Location:                      user.Location.String,
		DatingPreferences:             storeUserDatingPrefsToSlice(userDatingPrefs),
		DietaryRestrictions:           stringValuesToSettingCategories(dietaryRestrictions),
		DateTypePreferences:           stringValuesToSettingCategories(dateTypePreferences),
		MobilityConstraints:           stringValuesToSettingCategories(mobilityConstraints),
	}
}

func stringValuesToSettingCategories(values []string) []setting.Category {
	result := make([]setting.Category, len(values))
	for i, v := range values {
		result[i] = setting.Category{
			ID:   v, // String enum value used as ID
			Name: v,
		}
	}
	return result
}

func storeUserDatingPrefsToSlice(prefs pgmodel.UserDatingPreferenceSlice) []string {
	result := make([]string, 0, len(prefs))
	for _, pref := range prefs {
		result = append(result, pref.DatingPreference)
	}
	return result
}

// UpdatePersonalInformation updates the personal information of a user.
func (s *Store) UpdatePersonalInformation(ctx context.Context,
	exec boil.ContextExecutor,
	updater *setting.UpdatePersonalInformation,
) error {
	// update user
	if err := s.repoBackendApp.UpdateUser(ctx, newStoreUpdateUser(updater), exec); err != nil {
		return fmt.Errorf("wingedrepo.UpdateUser: %w", err)
	}

	// update user dating prefs
	for _, datePref := range updater.DatingPreferences {
		if err := s.repoBackendApp.UpsertUserDatingPreference(ctx, exec, updater.UserID, datePref); err != nil {
			return fmt.Errorf("wingedrepo.UpdateUser: %w", err)
		}
	}

	if updater.DietaryRestrictionIDs != nil {
		if _, err := s.repoBackendApp.SyncDietaryRestrictions(ctx, exec, &repo.SyncUserPreference{
			UserID: updater.UserID,
			Values: updater.DietaryRestrictionIDs,
		}); err != nil {
			return fmt.Errorf("sync dietary restrictions: %w", err)
		}
	}

	if updater.DateTypePreferenceIDs != nil {
		if _, err := s.repoBackendApp.SyncDateTypePreferences(ctx, exec, &repo.SyncUserPreference{
			UserID: updater.UserID,
			Values: updater.DateTypePreferenceIDs,
		}); err != nil {
			return fmt.Errorf("sync date type preferences: %w", err)
		}
	}

	if updater.MobilityConstraintIDs != nil {
		if _, err := s.repoBackendApp.SyncMobilityConstraints(ctx, exec, &repo.SyncUserPreference{
			UserID: updater.UserID,
			Values: updater.MobilityConstraintIDs,
		}); err != nil {
			return fmt.Errorf("sync mobility constraints: %w", err)
		}
	}

	return nil
}

func newStoreUpdateUser(updater *setting.UpdatePersonalInformation) *repo.UpdateUser {
	return &repo.UpdateUser{
		ID:                      updater.UserID,
		FirstName:               updater.Name,
		Birthday:                updater.Birthday,
		Number:                  updater.Number,
		Height:                  updater.Height,
		AgentDating:             updater.AgentDating,
		DatingPrefAgeRangeStart: updater.DatingPreferenceAgeRangeStart,
		DatingPrefAgeRangeEnd:   updater.DatingPreferenceAgeRangeEnd,
		Location:                updater.Location,
	}
}
