package extmatcher

import (
	"context"
	"testing"
	"wingedapp/pgtester/internal/util/strutil"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
	"wingedapp/pgtester/internal/wingedapp/lib/matching"

	"github.com/stretchr/testify/require"
)

func TestNewQualitativeMatcher(t *testing.T) {
	t.Skip() // I know this works, won't even bother hitting the external service every time
	q := NewQualitativeMatcher(applog.NewLogrus("test"))

	req := &matching.ProfileData{
		Romeo: &matching.PersonProfile{
			Qualitative: matching.QualitativeSection{
				SelfPortrait:               "string",
				Interests:                  "string",
				WellbeingHabits:            "string",
				SelfCareHabits:             "string",
				MoneyManagement:            "string",
				SelfReflectionCapabilities: "string",
				MoralFrameworks:            "string",
				LifeGoals:                  "string",
				PartnershipValues:          "string",
				MutualCommitment:           "string",
				SpiritualityGrowthMindset:  "string",
				CulturalValues:             "string",
				FamilyPlanning:             "string",
				IdealDate:                  "string",
				RedGreenFlags:              "string",
			},
			Quantitative: matching.QuantitativeSection{
				ExtroversionSocialEnergy: 1,
				RoutineVsSpontaneity:     1,
				Agreeableness:            1,
				Conscientiousness:        1,
				Neuroticism:              1,
				DominanceLevel:           1,
				EmotionalExpressiveness:  1,
				SexDrive:                 1,
				GeographicalMobility:     1,
			},
			Categorical: matching.CategoricalSection{
				ConflictResolutionStyle: "validating",
				SexualityPreferences:    "monogamy",
				Religion:                "christian",
			},
		},
		Juliet: &matching.PersonProfile{
			Qualitative: matching.QualitativeSection{
				SelfPortrait:               "string",
				Interests:                  "string",
				WellbeingHabits:            "string",
				SelfCareHabits:             "string",
				MoneyManagement:            "string",
				SelfReflectionCapabilities: "string",
				MoralFrameworks:            "string",
				LifeGoals:                  "string",
				PartnershipValues:          "string",
				MutualCommitment:           "string",
				SpiritualityGrowthMindset:  "string",
				CulturalValues:             "string",
				FamilyPlanning:             "string",
				IdealDate:                  "string",
				RedGreenFlags:              "string",
			},
			Quantitative: matching.QuantitativeSection{
				ExtroversionSocialEnergy: 1,
				RoutineVsSpontaneity:     1,
				Agreeableness:            1,
				Conscientiousness:        1,
				Neuroticism:              1,
				DominanceLevel:           1,
				EmotionalExpressiveness:  1,
				SexDrive:                 1,
				GeographicalMobility:     1,
			},
			Categorical: matching.CategoricalSection{
				ConflictResolutionStyle: "validating",
				SexualityPreferences:    "monogamy",
				Religion:                "christian",
			},
		},
	}

	compatRes, err := q.Qualify(context.Background(), req)
	require.NoError(t, err, "Qualify() should not return an error")
	require.NotNil(t, compatRes, "compatRes should not be nil")

	t.Log("===== compatRes:", strutil.GetAsJson(compatRes))
}
