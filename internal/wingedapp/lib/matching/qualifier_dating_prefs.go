package matching

import (
	"context"
	"errors"
	"fmt"
)

// newDatePrefsQualifier creates a dating preferences qualifier,
// and attaches it to the QualifierResults.
func newDatePrefsQualifier(qr *QualifierResults) *Qualifier {
	if qr.Distance != nil {
		panic(datePrefsQualifier + " already set")
	}
	qr.DatingPreferences = newQualifier(datePrefsQualifier)
	return qr.DatingPreferences
}

func (l *Logic) datePrefsQualifier(ctx context.Context, params *QualifierParameters) *Qualifier {
	q := newDatePrefsQualifier(params.QualifierResults)

	// guard: both users must have valid gender
	if !params.UserA.Gender.Valid || !params.UserB.Gender.Valid {
		return q.SetError(ErrGenderNotSet)
	}

	// guard: both users must have dating preferences
	if len(params.UserA.DatingPreferences) == 0 || len(params.UserB.DatingPreferences) == 0 {
		return q.SetError(ErrDatePrefsNotSet)
	}

	genderA := params.UserA.Gender.String
	genderB := params.UserB.Gender.String

	datePrefsA := params.UserA.DatingPreferences
	datePrefsB := params.UserB.DatingPreferences

	q.Telemetry["userA_gender"] = genderA
	q.Telemetry["userB_gender"] = genderB
	q.Telemetry["userA_date_prefs"] = datePrefsA
	q.Telemetry["userB_date_prefs"] = datePrefsB

	hashDatePrefsA := datingPrefsHash(datePrefsA)
	hashDatePrefsB := datingPrefsHash(datePrefsB)

	if ok := hashDatePrefsA[genderB]; !ok {
		return q.SetError(errors.Join(
			ErrNotInDatePrefs,
			fmt.Errorf("user A dating prefs do not include user B gender: %s", genderB),
		))
	}

	if ok := hashDatePrefsB[genderA]; !ok {
		return q.SetError(errors.Join(
			ErrNotInDatePrefs,
			fmt.Errorf("user B dating prefs do not include user A gender: %s", genderA),
		))
	}

	return q
}

// datingPrefsHash creates a hash map of dating preferences for quick lookup.
func datingPrefsHash(u []UserDatingPreference) map[string]bool {
	h := make(map[string]bool)
	for _, g := range u {
		h[g.DatingPreference] = true
	}
	return h
}
