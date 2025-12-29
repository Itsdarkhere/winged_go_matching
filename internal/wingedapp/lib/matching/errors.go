package matching

import "errors"

/* Matching logic sentinel errors */

var (
	ErrAgeGapMaleExceeds    = errors.New("male age gap exceeds limit")
	ErrAgeGapFemaleExceeds  = errors.New("female age gap exceeds limit")
	ErrAgeGapSameSexExceeds = errors.New("same sex age gap exceeds limit")
	ErrDistanceExceeds      = errors.New("distance exceeds limit")
	ErrNotInDatePrefs       = errors.New("not in date preferences")
	ErrHeightGapExists      = errors.New("height gap exists")

	ErrAgeGapHetero = errors.New("hetero age gap too large")
	ErrNoMale       = errors.New("no male user")
	ErrNoFemale     = errors.New("no female user")

	// null field validation errors
	ErrAgeNotSet       = errors.New("age not set for one or both users")
	ErrGenderNotSet    = errors.New("gender not set for one or both users")
	ErrHeightNotSet    = errors.New("height not set for one or both users")
	ErrLocationNotSet  = errors.New("location not set for one or both users")
	ErrDatePrefsNotSet = errors.New("dating preferences not set for one or both users")

	// config update validation errors
	ErrAdaptiveExpansionNotIncrementing = errors.New("location_adaptive_expansion must be a strictly incrementing array (e.g., [10, 20, 30])")
	ErrAgeRangeStartGreaterThanEnd      = errors.New("age_range_start must be less than or equal to age_range_end")
	ErrScoreRangeStartGreaterThanEnd    = errors.New("score_range_start must be less than or equal to score_range_end")
	ErrDropHoursInvalidFormat           = errors.New("drop_hours must contain valid time strings in HH:MM format (e.g., \"19:00\")")
	ErrDropHoursUTCInvalidFormat        = errors.New("drop_hours_utc must contain valid timezone strings (e.g., \"GMT+3\")")
	ErrNegativeValue                    = errors.New("numeric configuration values must be non-negative")
	ErrConfigNotFound                   = errors.New("match configuration not found")
)
