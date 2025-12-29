package matching

import (
	"context"
	"errors"
	"fmt"
	"math"
)

// newAgeWindowQualifier creates a new age window qualifier,
// and attaches it to the QualifierResults.
func newAgeWindowQualifier(qr *QualifierResults) *Qualifier {
	if qr.AgeWindow != nil {
		panic(ageWindowQualifier + " already set")
	}
	qr.AgeWindow = newQualifier(ageWindowQualifier)
	return qr.AgeWindow
}

// ageWindowQualifier checks if the age gap between two users is acceptable
func (l *Logic) ageQualifier(ctx context.Context, params *QualifierParameters) *Qualifier {
	q := newAgeWindowQualifier(params.QualifierResults)

	userA := params.UserA
	userB := params.UserB

	// guard: both users must have valid age
	if !userA.Age.Valid || !userB.Age.Valid {
		return q.SetError(ErrAgeNotSet)
	}

	// guard: both users must have valid gender for hetero check
	if !userA.Gender.Valid || !userB.Gender.Valid {
		return q.SetError(ErrGenderNotSet)
	}

	// base case: same age - all good
	if userA.Age.Int == userB.Age.Int {
		return q
	}

	// PRIMARY LOGIC: hetero, and non-hetero couples
	// have different age gap rules.
	isHeteroMatch := usersAreHetero(userA, userB)

	ageGap := math.Abs(float64(userA.Age.Int - userB.Age.Int))
	cfg := params.config

	q.Telemetry["isHeteroMatch"] = isHeteroMatch
	q.Telemetry["ageGap"] = ageGap

	if isHeteroMatch {
		maleUser, femaleUser, err := getMaleAndFemaleUsers(userA, userB)
		if err != nil {
			return q.SetError(err)
		}

		q.Telemetry["maleUserAge"] = maleUser.Age.Int
		q.Telemetry["femaleUserAge"] = femaleUser.Age.Int

		// male older
		if (maleUser.Age.Int > femaleUser.Age.Int) && (ageGap > float64(cfg.AgeRangeManOlderBy)) {
			return q.SetError(errors.Join(
				ErrAgeGapMaleExceeds,
				fmt.Errorf("m age: %d, f age: %d, max allowed gap: %d",
					maleUser.Age.Int, femaleUser.Age.Int, cfg.AgeRangeManOlderBy,
				),
			))
		}

		// female older
		if (femaleUser.Age.Int > maleUser.Age.Int) && (ageGap > float64(cfg.AgeRangeWomanOlderBy)) {
			return q.SetError(errors.Join(
				ErrAgeGapFemaleExceeds,
				fmt.Errorf("m age: %d, f age: %d, max allowed gap: %d",
					maleUser.Age.Int, femaleUser.Age.Int, cfg.AgeRangeManOlderBy,
				),
			))
		}

		return q
	}

	// same sex check
	if ageGap > float64(cfg.AgeRangeEnd) {
		return q.SetError(errors.Join(
			ErrAgeGapSameSexExceeds,
			fmt.Errorf("user-a age: %d, user-b age: %d, max allowed gap: %d",
				userA.Age.Int, userB.Age.Int, cfg.AgeRangeEnd,
			),
		))
	}

	return q
}

func usersAreHetero(userA *User, userB *User) bool {
	// both genders must be valid to determine hetero status
	if !userA.Gender.Valid || !userB.Gender.Valid {
		return false
	}

	isDifferentGenders := userA.Gender.String != userB.Gender.String
	oneIsMale := userA.Gender.String == male || userB.Gender.String == male
	oneIsFemale := userA.Gender.String == female || userB.Gender.String == female

	return isDifferentGenders && oneIsMale && oneIsFemale
}

// getMaleAndFemaleUsers returns the male and female users from the given two users.
func getMaleAndFemaleUsers(userA *User, userB *User) (*User, *User, error) {
	maleUser, err := getMaleUser(userA, userB)
	if err != nil {
		return nil, nil, err
	}

	femaleUser, err := getFemaleUser(userA, userB)
	if err != nil {
		return nil, nil, err
	}

	return maleUser, femaleUser, nil
}
