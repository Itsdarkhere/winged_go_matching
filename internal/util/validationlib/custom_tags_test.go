package validationlib

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFutureTimeValidation(t *testing.T) {
	// Find the future_time validator
	var futureTimeFunc func(interface{}) bool
	for _, cv := range customValidations {
		if cv.Name == "future_time" {
			futureTimeFunc = cv.Logic.Func
			break
		}
	}

	if futureTimeFunc == nil {
		t.Fatal("future_time validation not found")
	}

	testCases := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{
			name:     "success-future-time",
			input:    time.Now().Add(24 * time.Hour),
			expected: true,
		},
		{
			name:     "error-past-time",
			input:    time.Now().Add(-1 * time.Hour),
			expected: false,
		},
		{
			name:     "error-current-time-boundary",
			input:    time.Now().Add(-1 * time.Second),
			expected: false,
		},
		{
			name:     "error-wrong-type",
			input:    "not a time",
			expected: false,
		},
		{
			name:     "error-nil-value",
			input:    nil,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := futureTimeFunc(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFutureTimeMin2HValidation(t *testing.T) {
	// Find the future_time_min_2h validator
	var futureTimeMin2HFunc func(interface{}) bool
	for _, cv := range customValidations {
		if cv.Name == "future_time_min_2h" {
			futureTimeMin2HFunc = cv.Logic.Func
			break
		}
	}

	if futureTimeMin2HFunc == nil {
		t.Fatal("future_time_min_2h validation not found")
	}

	testCases := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{
			name:     "success-3-hours-ahead",
			input:    time.Now().Add(3 * time.Hour),
			expected: true,
		},
		{
			name:     "success-24-hours-ahead",
			input:    time.Now().Add(24 * time.Hour),
			expected: true,
		},
		{
			name:     "error-1-hour-ahead",
			input:    time.Now().Add(1 * time.Hour),
			expected: false,
		},
		{
			name:     "error-1h59m-ahead",
			input:    time.Now().Add(1*time.Hour + 59*time.Minute),
			expected: false,
		},
		{
			name:     "error-past-time",
			input:    time.Now().Add(-1 * time.Hour),
			expected: false,
		},
		{
			name:     "error-wrong-type",
			input:    "not a time",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := futureTimeMin2HFunc(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestTimeSliceAllFutureValidation(t *testing.T) {
	// Find the time_slice_all_future validator
	var timeSliceAllFutureFunc func(interface{}) bool
	for _, cv := range customValidations {
		if cv.Name == "time_slice_all_future" {
			timeSliceAllFutureFunc = cv.Logic.Func
			break
		}
	}

	if timeSliceAllFutureFunc == nil {
		t.Fatal("time_slice_all_future validation not found")
	}

	now := time.Now()
	future1 := now.Add(24 * time.Hour)
	future2 := now.Add(48 * time.Hour)
	past := now.Add(-1 * time.Hour)

	testCases := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{
			name:     "success-all-future",
			input:    []time.Time{future1, future2},
			expected: true,
		},
		{
			name:     "success-single-future",
			input:    []time.Time{future1},
			expected: true,
		},
		{
			name:     "success-empty-slice",
			input:    []time.Time{},
			expected: true,
		},
		{
			name:     "error-one-past",
			input:    []time.Time{future1, past},
			expected: false,
		},
		{
			name:     "error-all-past",
			input:    []time.Time{past, past},
			expected: false,
		},
		{
			name:     "error-wrong-type",
			input:    "not a slice",
			expected: false,
		},
		{
			name:     "error-wrong-slice-type",
			input:    []string{"2025-01-01"},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := timeSliceAllFutureFunc(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestTimeSliceNoDuplicatesValidation(t *testing.T) {
	// Find the time_slice_no_duplicates validator
	var timeSliceNoDuplicatesFunc func(interface{}) bool
	for _, cv := range customValidations {
		if cv.Name == "time_slice_no_duplicates" {
			timeSliceNoDuplicatesFunc = cv.Logic.Func
			break
		}
	}

	if timeSliceNoDuplicatesFunc == nil {
		t.Fatal("time_slice_no_duplicates validation not found")
	}

	time1 := time.Date(2025, 12, 10, 19, 0, 0, 0, time.UTC)
	time2 := time.Date(2025, 12, 11, 19, 0, 0, 0, time.UTC)
	time3 := time.Date(2025, 12, 12, 19, 0, 0, 0, time.UTC)

	testCases := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{
			name:     "success-no-duplicates",
			input:    []time.Time{time1, time2},
			expected: true,
		},
		{
			name:     "success-three-unique",
			input:    []time.Time{time1, time2, time3},
			expected: true,
		},
		{
			name:     "success-single-time",
			input:    []time.Time{time1},
			expected: true,
		},
		{
			name:     "success-empty-slice",
			input:    []time.Time{},
			expected: true,
		},
		{
			name:     "error-has-duplicates",
			input:    []time.Time{time1, time1},
			expected: false,
		},
		{
			name:     "error-has-duplicates-in-middle",
			input:    []time.Time{time1, time2, time2, time3},
			expected: false,
		},
		{
			name:     "error-wrong-type",
			input:    "not a slice",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := timeSliceNoDuplicatesFunc(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
