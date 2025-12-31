package store

import (
	"context"
	"fmt"
	"time"
	"wingedapp/pgtester/internal/wingedapp/aibackend/db/aipgmodel"
	"wingedapp/pgtester/internal/wingedapp/aibackend/db/repo"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
	"wingedapp/pgtester/internal/wingedapp/lib/matching"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
)

func NewProfileStore(l applog.Logger) *ProfileStore {
	return &ProfileStore{
		l:    l,
		repo: repo.NewStore(),
	}
}

type ProfileStore struct {
	l    applog.Logger
	repo *repo.Store
}

// Insert inserts a profile into ai_backend.profiles.
func (p *ProfileStore) Insert(
	ctx context.Context,
	exec boil.ContextExecutor,
	params *matching.InsertPopulationProfile,
) error {
	if params == nil || params.Details == nil {
		return fmt.Errorf("params and details cannot be nil")
	}

	d := params.Details
	profile := &aipgmodel.Profile{
		ID:        uuid.NewString(),
		UserID:    null.StringFrom(params.UserID),
		CreatedAt: time.Now(),
		UpdatedAt: null.TimeFrom(time.Now()),

		// Qualitative fields
		WellbeingHabits:            d.WellbeingHabits,
		CorePassions:               d.Interests,
		SenseOfHumor:               d.SenseOfHumor,
		SelfCareHabits:             d.SelfCareHabits,
		MoneyManagement:            d.MoneyManagement,
		SelfReflectionCapabilities: d.SelfReflectionCapabilities,
		MoralFrameworks:            d.MoralFrameworks,
		LifeGoals:                  d.LifeGoals,
		PartnershipValues:          d.PartnershipValues,
		SpiritualityGrowthMindset:  d.SpiritualityGrowthMindset,
		CulturalValues:             d.CulturalValues,
		FamilyPlanning:             d.FamilyPlanning,
		SelfPortrait:               d.SelfPortrait,
		MutualCommitment:           d.MutualCommitment,
		IdealDate:                  d.IdealDate,
		RedGreenFlags:              d.RedGreenFlags,
		CasualInterests:            d.CasualInterests,
		CommunicationStyle:         d.CommunicationStyle,
		LifeSituation:              d.LifeSituation,

		// Quantitative fields
		ExtroversionSocialEnergy: d.ExtroversionSocialEnergy,
		RoutineVSSpontaneity:     d.RoutineVsSpontaneity,
		Agreeableness:            d.Agreeableness,
		Conscientiousness:        d.Conscientiousness,
		DominanceLevel:           d.DominanceLevel,
		EmotionalExpressiveness:  d.EmotionalExpressiveness,
		SexDrive:                 d.SexDrive,
		GeographicalMobility:     d.GeographicalMobility,
		Neuroticism:              d.Neuroticism,
		FinancialRiskTolerance:   d.FinancialRiskTolerance,
		AbstractVsConcrete:       d.AbstractVsConcrete,

		// Categorical fields
		ConflictResolutionStyle: d.ConflictResolutionStyle,
		SexualityPreferences:    d.SexualityPreferences,
		Religion:                d.Religion,
		AttachmentStyle:         d.AttachmentStyle,
	}

	if err := p.repo.InsertProfile(ctx, exec, profile); err != nil {
		return fmt.Errorf("insert profile: %w", err)
	}

	return nil
}

func (p *ProfileStore) Profile(ctx context.Context, exec boil.ContextExecutor, userUUID uuid.UUID) (*matching.PersonProfile, error) {
	pgProfile, err := p.repo.Profile(ctx, exec, &repo.ProfileQueryFilter{
		UserID: null.StringFrom(userUUID.String()),
	})
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}

	if pgProfile == nil {
		return nil, nil
	}

	profile := &matching.PersonProfile{
		Qualitative: matching.QualitativeSection{
			SelfPortrait:               pgProfile.SelfPortrait.String,
			CorePassions:               pgProfile.CorePassions.String,
			WellbeingHabits:            pgProfile.WellbeingHabits.String,
			SelfCareHabits:             pgProfile.SelfCareHabits.String,
			MoneyManagement:            pgProfile.MoneyManagement.String,
			SelfReflectionCapabilities: pgProfile.SelfReflectionCapabilities.String,
			MoralFrameworks:            pgProfile.MoralFrameworks.String,
			LifeGoals:                  pgProfile.LifeGoals.String,
			PartnershipValues:          pgProfile.PartnershipValues.String,
			MutualCommitment:           pgProfile.MutualCommitment.String,
			SpiritualityGrowthMindset:  pgProfile.SpiritualityGrowthMindset.String,
			CulturalValues:             pgProfile.CulturalValues.String,
			FamilyPlanning:             pgProfile.FamilyPlanning.String,
			IdealDate:                  pgProfile.IdealDate.String,
			RedGreenFlags:              pgProfile.RedGreenFlags.String,
			SenseOfHumor:               pgProfile.SenseOfHumor.String,
			CasualInterests:            pgProfile.CasualInterests.String,
			CommunicationStyle:         pgProfile.CommunicationStyle.String,
			LifeSituation:              pgProfile.LifeSituation.String,
		},
		Quantitative: matching.QuantitativeSection{
			ExtroversionSocialEnergy: pgProfile.ExtroversionSocialEnergy.Float64,
			RoutineVsSpontaneity:     pgProfile.RoutineVSSpontaneity.Float64,
			Agreeableness:            pgProfile.Agreeableness.Float64,
			Conscientiousness:        pgProfile.Conscientiousness.Float64,
			Neuroticism:              float64(pgProfile.Neuroticism.Float32),
			DominanceLevel:           pgProfile.DominanceLevel.Float64,
			EmotionalExpressiveness:  pgProfile.EmotionalExpressiveness.Float64,
			SexDrive:                 pgProfile.SexDrive.Float64,
			GeographicalMobility:     pgProfile.GeographicalMobility.Float64,
			FinancialRiskTolerance:   pgProfile.FinancialRiskTolerance.Float64,
			AbstractVsConcrete:       pgProfile.AbstractVsConcrete.Float64,
		},
		Categorical: matching.CategoricalSection{
			ConflictResolutionStyle: pgProfile.ConflictResolutionStyle.String,
			SexualityPreferences:    pgProfile.SexualityPreferences.String,
			Religion:                pgProfile.Religion.String,
			AttachmentStyle:         pgProfile.AttachmentStyle.String,
		},
	}

	return profile, nil
}

// DeleteByUserIDs deletes profiles by user IDs.
// Returns the number of profiles deleted.
func (p *ProfileStore) DeleteByUserIDs(
	ctx context.Context,
	exec boil.ContextExecutor,
	userIDs []string,
) (int64, error) {
	if len(userIDs) == 0 {
		return 0, nil
	}

	deleted, err := aipgmodel.Profiles(
		aipgmodel.ProfileWhere.UserID.IN(userIDs),
	).DeleteAll(ctx, exec)
	if err != nil {
		return 0, fmt.Errorf("delete profiles: %w", err)
	}

	p.l.Debug(ctx, "deleted ai_backend profiles by user_id", applog.F("deleted_count", deleted), applog.F("user_ids", userIDs))

	return deleted, nil
}
