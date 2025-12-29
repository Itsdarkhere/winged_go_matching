package pg

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/aibackend/db/aipgmodel"
	"wingedapp/pgtester/internal/wingedapp/aibackend/db/repo"
	"wingedapp/pgtester/internal/wingedapp/business/domain/registration"

	"github.com/aarondl/sqlboiler/v4/boil"
)

// Profiles retrieves a list of profiles based on the provided filter.
func (s *Store) Profiles(ctx context.Context, exec boil.ContextExecutor) ([]registration.Profile, error) {
	pgProfiles, err := s.repoAIBackend.Profiles(ctx,
		exec,
		&repo.ProfileQueryFilter{},
	)
	if err != nil {
		return nil, fmt.Errorf("list profiles: %w", err)
	}

	return newProfilesFromSlice(pgProfiles), nil
}

// Profile gets details of a specific profiles based on the provided filter.
func (s *Store) Profile(ctx context.Context, exec boil.ContextExecutor, filter *registration.ProfileQueryFilter) (*registration.Profile, error) {
	profiles, err := s.repoAIBackend.Profiles(ctx, exec, &repo.ProfileQueryFilter{
		ID:     filter.ID,
		UserID: filter.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("list profiles: %w", err)
	}
	if len(profiles) == 0 {
		return nil, registration.ErrProfileNotFound
	}
	if len(profiles) != 1 {
		return nil, fmt.Errorf("profile count mismatch, have %d, want 1", len(profiles))
	}

	newProfile := newProfilesFromSlice(profiles)

	return &newProfile[0], nil
}

func newProfilesFromSlice(pgProfiles aipgmodel.ProfileSlice) []registration.Profile {
	if pgProfiles == nil {
		return nil
	}

	profiles := make([]registration.Profile, 0, len(pgProfiles))
	for _, pgProfile := range pgProfiles {
		profiles = append(profiles, registration.Profile{
			ID:                         pgProfile.ID,
			UserID:                     pgProfile.UserID,
			ConversationID:             pgProfile.ConversationID,
			WellbeingHabits:            pgProfile.WellbeingHabits,
			Interests:                  pgProfile.Interests,
			SenseOfHumor:               pgProfile.SenseOfHumor,
			SelfCareHabits:             pgProfile.SelfCareHabits,
			MoneyManagement:            pgProfile.MoneyManagement,
			SelfReflectionCapabilities: pgProfile.SelfReflectionCapabilities,
			MoralFrameworks:            pgProfile.MoralFrameworks,
			LifeGoals:                  pgProfile.LifeGoals,
			PartnershipValues:          pgProfile.PartnershipValues,
			SpiritualityGrowthMindset:  pgProfile.SpiritualityGrowthMindset,
			CulturalValues:             pgProfile.CulturalValues,
			FamilyPlanning:             pgProfile.FamilyPlanning,
			ExtroversionSocialEnergy:   pgProfile.ExtroversionSocialEnergy,
			RoutineVSSpontaneity:       pgProfile.RoutineVSSpontaneity,
			Agreeableness:              pgProfile.Agreeableness,
			Conscientiousness:          pgProfile.Conscientiousness,
			DominanceLevel:             pgProfile.DominanceLevel,
			EmotionalExpressiveness:    pgProfile.EmotionalExpressiveness,
			SexDrive:                   pgProfile.SexDrive,
			GeographicalMobility:       pgProfile.GeographicalMobility,
			ConflictResolutionStyle:    pgProfile.ConflictResolutionStyle,
			SexualityPreferences:       pgProfile.SexualityPreferences,
			Religion:                   pgProfile.Religion,
			SelfPortrait:               pgProfile.SelfPortrait,
			MutualCommitment:           pgProfile.MutualCommitment,
			IdealDate:                  pgProfile.IdealDate,
			RedGreenFlags:              pgProfile.RedGreenFlags,
			Neuroticism:                pgProfile.Neuroticism,
		})
	}

	return profiles
}
