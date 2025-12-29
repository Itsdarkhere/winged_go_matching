package timerange

import (
	"github.com/golang-module/carbon/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type testCaseOverlaps struct {
	name       string
	spanPoint1 Span
	spanPoint2 Span
	asserts    func(t *testing.T, res *Span)
}

func overlapTestCases() []testCaseOverlaps {
	now := carbon.Now()

	return []testCaseOverlaps{
		{
			name: "Overlapping spans",
			spanPoint1: Span{
				Start: now,
				End:   now.AddHours(5),
			},
			spanPoint2: Span{
				Start: now.AddHours(2),
				End:   now.AddHours(6),
			},
			asserts: func(t *testing.T, res *Span) {
				require.NotNil(t, res, "Expected overlap span to be not nil")
				diffInHours := res.Start.DiffInHours(res.End)
				require.Equal(t, 3, int(diffInHours), "Expected overlap span duration to be 3 hours")
			},
		},
		{
			name: "Non-overlapping spans",
			spanPoint1: Span{
				Start: now.AddYears(1),
				End:   now.AddYears(1).AddHours(5),
			},
			spanPoint2: Span{
				Start: now,
				End:   now.AddHours(6),
			},
			asserts: func(t *testing.T, res *Span) {
				require.Nil(t, res, "Expected overlap span to be nil")
			},
		},
		{
			name: "Touching ends",
			spanPoint1: Span{
				Start: now,
				End:   now.AddHours(1),
			},
			spanPoint2: Span{
				Start: now.AddHours(1),
				End:   now.AddHours(2),
			},
			asserts: func(t *testing.T, res *Span) {
				require.Nil(t, res, "Expected overlap span to be nil")
			},
		},

		// business case from spec
		{
			name: "biz case 1 (13-20) (14-21) (14-20)",
			spanPoint1: Span{
				Start: now.SetHour(13),
				End:   now.SetHour(20),
			},
			spanPoint2: Span{
				Start: now.SetHour(14),
				End:   now.SetHour(21),
			},
			asserts: func(t *testing.T, res *Span) {
				require.NotNil(t, res, "Expected overlap span to be not nil")

				assert.Equal(t, 14, res.Start.Hour(), "Expected overlap start Hour to be 14")
				assert.Equal(t, 20, res.End.Hour(), "Expected overlap end Hour to be 20")
			},
		},
		{
			name: "biz case 2 (13:30-15) (9-14) (13:30-14)",
			spanPoint1: Span{
				Start: now.SetHour(13).SetMinute(30),
				End:   now.SetHour(15),
			},
			spanPoint2: Span{
				Start: now.SetHour(9),
				End:   now.SetHour(14),
			},
			asserts: func(t *testing.T, res *Span) {
				require.NotNil(t, res, "Expected overlap span to be not nil")

				assert.Equal(t, 13, res.Start.Hour(), "Expected overlap start Hour to be 13")
				assert.Equal(t, 30, res.Start.Minute(), "Expected overlap start minute to be 30")
				assert.Equal(t, 14, res.End.Hour(), "Expected overlap end Hour to be 14")
			},
		},
		{
			name: "biz case 3 (16-01) (17-00:30) (17-00:30)",
			spanPoint1: Span{
				Start: now.SetHour(16).SetMinute(30),
				End:   now.AddDays(1).SetHour(1),
			},
			spanPoint2: Span{
				Start: now.SetHour(17),
				End:   now.AddDays(1).SetHour(0).SetMinute(30),
			},
			asserts: func(t *testing.T, res *Span) {
				require.NotNil(t, res, "Expected overlap span to be not nil")

				assert.Equal(t, 17, res.Start.Hour(), "Expected overlap start Hour to be 17")
				assert.Equal(t, 0, res.End.Hour(), "Expected overlap end Hour to be 0")
				assert.Equal(t, 30, res.End.Minute(), "Expected overlap end minute to be 30")
			},
		},
		{
			name: "biz case 4 (13-15) (09-13) (Closed)",
			spanPoint1: Span{
				Start: now.SetHour(13),
				End:   now.SetHour(15),
			},
			spanPoint2: Span{
				Start: now.SetHour(9),
				End:   now.SetHour(13),
			},
			asserts: func(t *testing.T, res *Span) {
				require.Nil(t, res, "Expected overlap span to be nil")
			},
		},
		{
			name: "biz case 5 (13-15) (14:31-16)",
			spanPoint1: Span{
				Start: now.SetHour(13),
				End:   now.SetHour(15),
			},
			spanPoint2: Span{
				Start: now.SetHour(14).SetMinute(31),
				End:   now.SetHour(16),
			},
			asserts: func(t *testing.T, res *Span) {
				require.NotNil(t, res, "Expected overlap span to be not nil")
			},
		},
	}
}

func Test_Overlaps(t *testing.T) {
	for _, tc := range overlapTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			o := OverlapSpan(tc.spanPoint1, tc.spanPoint2)
			tc.asserts(t, o)
		})
	}
}

type testCaseOverlappingSpans struct {
	name    string
	spans1  []Span
	spans2  []Span
	asserts func(t *testing.T, res []Span, err error)
}

func overlappingRangesTestCases() []testCaseOverlappingSpans {
	now := carbon.Now().SetHour(0).SetMinute(0)

	return []testCaseOverlappingSpans{
		{
			name: "error has overlap in spans1",
			spans1: []Span{
				{
					Start: now.AddHours(1),
					End:   now.AddHours(6),
				},
				{
					Start: now.AddHours(2),
					End:   now.AddHours(6),
				},
			},
			spans2: []Span{
				{
					Start: now.AddHours(2),
					End:   now.AddHours(4),
				},
			},
			asserts: func(t *testing.T, res []Span, err error) {
				require.Nil(t, res, "Expected overlap span to be not nil")
				require.Error(t, err, "Expected error due to overlap in spans1")
				require.ErrorContains(t, err, ErrorSpansAHasOverlap.Error())
			},
		},
		{
			name: "error has overlap in spans2",
			spans1: []Span{
				{
					Start: now.AddHours(1),
					End:   now.AddHours(6),
				},
			},
			spans2: []Span{
				{
					Start: now.AddHours(1),
					End:   now.AddHours(4),
				},
				{
					Start: now.AddHours(2),
					End:   now.AddHours(4),
				},
			},
			asserts: func(t *testing.T, res []Span, err error) {
				require.Nil(t, res, "Expected overlap span to be not nil")
				require.Error(t, err, "Expected error due to overlap in spans2")
				require.ErrorContains(t, err, ErrorSpansBHasOverlap.Error())
			},
		},
		{
			name: "Overlapping spans",
			spans1: []Span{
				{
					Start: now.AddHours(1),
					End:   now.AddHours(6),
				},
			},
			spans2: []Span{
				{
					Start: now.AddHours(2),
					End:   now.AddHours(4),
				},
			},
			asserts: func(t *testing.T, res []Span, err error) {
				require.NoError(t, err, "Expected no error for overlapping spans")
				require.NotNil(t, res, "Expected overlap span to be not nil")

				// I will implicitly also test more granularly in the biz instead
			},
		},
	}
}

func Test_OverlappingSpans(t *testing.T) {
	for _, tc := range overlappingRangesTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			o, err := OverlappingSpans(tc.spans1, tc.spans2)
			tc.asserts(t, o, err)
		})
	}
}
