package matching

import (
	"context"
	"errors"
	"fmt"
	"math"
)

// newHeightQualifier creates a new distance qualifier,
// and attaches it to the QualifierResults.
func newHeightQualifier(qr *QualifierResults) *Qualifier {
	if qr.Height != nil {
		panic(heightQualifier + " already set")
	}
	qr.Height = newQualifier(heightQualifier)
	return qr.Height
}

// distanceQualifier checks if the distance between two users is acceptable.
func (l *Logic) heightQualifier(ctx context.Context, params *QualifierParameters) *Qualifier {
	q := newHeightQualifier(params.QualifierResults)

	// guard: both users must have valid height
	if !params.UserA.Height.Valid || !params.UserB.Height.Valid {
		return q.SetError(ErrHeightNotSet)
	}

	// guard: both users must have valid gender
	if !params.UserA.Gender.Valid || !params.UserB.Gender.Valid {
		return q.SetError(ErrGenderNotSet)
	}

	heightA := params.UserA.Height.Float64
	heightB := params.UserB.Height.Float64
	heightDiff := math.Abs(heightB - heightA)

	hasNonBinary := oneIsNonBinary(params.UserA, params.UserB)
	sameSex := isSameSex(params.UserA, params.UserB)

	q.Telemetry["userA_height_cm"] = heightA
	q.Telemetry["userB_height_cm"] = heightB
	q.Telemetry["has_non_binary"] = hasNonBinary
	q.Telemetry["same_sex"] = sameSex
	q.Telemetry["height_difference_cm"] = heightDiff

	if hasNonBinary || sameSex {
		q.Telemetry["relaxed_height_preference"] = "relaxing due to same sex, or non-binary user"
		return q // relax height preference
	}

	// else, enforce height preference
	maleUser, femaleUser, err := getMaleAndFemaleUsers(params.UserA, params.UserB)
	if err != nil {
		return q.SetError(err)
	}

	q.Telemetry["male_user_height"] = maleUser.Height.Float64
	q.Telemetry["female_user_height"] = femaleUser.Height.Float64
	q.Telemetry["female_user_height_with_allowance"] = femaleUser.Height.Float64 + params.config.HeightMaleGreaterByCM

	// main formula: male a little taller
	if maleUser.Height.Float64 >= (femaleUser.Height.Float64 + params.config.HeightMaleGreaterByCM) {
		return q // all good
	}

	return q.SetError(errors.Join(
		ErrHeightGapExists,
		fmt.Errorf("male height %v, female height %v, required gap %v cm",
			maleUser.Height.Float64,
			femaleUser.Height.Float64,
			params.config.HeightMaleGreaterByCM,
		),
	))
}

func oneIsNonBinary(userA, userB *User) bool {
	users := []*User{userA, userB}
	for _, user := range users {
		if user.Gender.Valid && user.Gender.String == nonBinary {
			return true
		}
	}

	return false
}

func isSameSex(a, b *User) bool {
	// both genders must be valid to determine same sex
	if !a.Gender.Valid || !b.Gender.Valid {
		return false
	}
	return a.Gender.String == b.Gender.String
}
