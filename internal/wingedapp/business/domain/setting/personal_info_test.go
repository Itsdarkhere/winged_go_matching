package setting

import (
	"testing"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func validUpdatePersInfo() *UpdatePersonalInformation {
	return &UpdatePersonalInformation{
		UserID:                        uuid.NewString(),
		Name:                          null.StringFrom("test"),
		Birthday:                      null.TimeFrom(time.Now()),
		Number:                        null.StringFrom("1234567890"),
		Height:                        null.IntFrom(170),
		AgentDating:                   null.BoolFrom(true),
		DatingPreferences:             []string{datingPrefMale, datingPrefFemale},
		DatingPreferenceAgeRangeStart: null.IntFrom(18),
		DatingPreferenceAgeRangeEnd:   null.IntFrom(20), // valid because end > start
		Location:                      null.StringFrom("Helsinki, FI"),
		updateDatingPrefs:             true,
	}
}

func TestBusiness_collectUpdatePersInfoErrors(t *testing.T) {
	tests := []struct {
		name       string
		mutate     func(u *UpdatePersonalInformation)
		expectErrs []string
		wantErr    bool
	}{
		{
			name:    "valid input",
			mutate:  func(u *UpdatePersonalInformation) {},
			wantErr: false,
		},
		{
			name: "empty name string with Valid=true",
			mutate: func(u *UpdatePersonalInformation) {
				u.Name = null.StringFrom("")
			},
			expectErrs: []string{"name is empty"},
			wantErr:    true,
		},
		{
			name: "zero birthday with Valid=true",
			mutate: func(u *UpdatePersonalInformation) {
				u.Birthday = null.TimeFrom(time.Time{})
			},
			expectErrs: []string{"birthday is empty"},
			wantErr:    true,
		},
		{
			name: "empty number string with Valid=true",
			mutate: func(u *UpdatePersonalInformation) {
				u.Number = null.StringFrom("")
			},
			expectErrs: []string{"number is empty"},
			wantErr:    true,
		},
		{
			name: "zero height with Valid=true",
			mutate: func(u *UpdatePersonalInformation) {
				u.Height = null.IntFrom(0)
			},
			expectErrs: []string{"height is empty"},
			wantErr:    true,
		},
		{
			name: "empty location string with Valid=true",
			mutate: func(u *UpdatePersonalInformation) {
				u.Location = null.StringFrom("")
			},
			expectErrs: []string{"location is empty"},
			wantErr:    true,
		},
		{
			name: "age range start greater than end",
			mutate: func(u *UpdatePersonalInformation) {
				u.DatingPreferenceAgeRangeStart = null.IntFrom(30)
				u.DatingPreferenceAgeRangeEnd = null.IntFrom(25) // invalid
			},
			expectErrs: []string{"date pref has errors", ErrDatePrefAgeEndGtStart.Error()},
			wantErr:    true,
		},
		{
			name: "age range start below minimum age",
			mutate: func(u *UpdatePersonalInformation) {
				u.DatingPreferenceAgeRangeStart = null.IntFrom(14) // < 18
				u.DatingPreferenceAgeRangeEnd = null.IntFrom(15)
			},
			expectErrs: []string{"date pref has errors", ErrBelowMinAgeRange.Error()},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			u := validUpdatePersInfo()
			tt.mutate(u)

			biz := &Business{cfg: &Config{ThresholdMinimumAge: 18}}
			errs := biz.collectUpdatePersInfoErrors(u)

			if tt.wantErr {
				require.True(t, errs.HasErrors(), "expected errors but got none")
				errStr := errs.Error().Error()
				for _, exp := range tt.expectErrs {
					require.Contains(t, errStr, exp)
				}
			} else {
				require.False(t, errs.HasErrors(), "expected no errors but got: %v", errs)
			}
		})
	}
}
