package testhelper

const (
	mockSuccessPopulationResponse = `
{
    "personality_compatibility_score": {
        "score": 1,
        "explanation": "Both Romeo and Juliet have identical high scores (1.0) in extroversion_social_energy, agreeableness, conscientiousness, neuroticism, dominance_level, emotional_expressiveness, and routine_vs_spontaneity, indicating no potential personality mismatches and high alignment. Both have the same conflict_resolution_style as 'validating' and reported self_reflection_capabilities as 'string', indicating presumed reflective abilities without conflict."
    },
    "lifestyle_compatibility_score": {
        "score": 0.6,
        "explanation": "Lifestyle fields such as interests, wellbeing_habits, self_care_habits, money_management, and typical social activities are unknown (all labeled 'string'), leading to a mild penalty, shrinking the score toward 0.60. Geographical mobility is identical at 1.0 but is ignored per rubric. No known shared activities or differences to enhance or penalize further."
    },
    "values_compatibility_score": {
        "score": 0.6,
        "explanation": "Core values including moral_frameworks, friendship_values, mutual_support, spirituality_growth_mindset, cultural_values, life_goals, and red_green_flags are all unknown (not provided, represented by 'string'). This unknown status reduces confidence in value compatibility and thus lowers the values compatibility score to 0.60 without positive or negative evidence."
    },
    "total_score": 0.73
}

`
)
