package timerange

import (
	"errors"
	"fmt"
	"sort"

	"github.com/golang-module/carbon/v2"
)

type Span struct {
	Start carbon.Carbon
	End   carbon.Carbon
}

func overlaps(spanPoint1 Span, spanPoint2 Span) bool {
	if spanPoint1.Start.Gt(spanPoint2.Start) && spanPoint1.Start.Lt(spanPoint2.End) {
		return true
	}

	if spanPoint1.End.Gt(spanPoint2.Start) && spanPoint1.End.Lt(spanPoint2.End) {
		return true
	}

	return false
}

// Overlaps checks if two time spans overlap.
// Checks if start and end of spanPoint1 are between spanPoint2
func Overlaps(spanPoint1 Span, spanPoint2 Span) bool {
	if overlaps(spanPoint1, spanPoint2) {
		return true
	}

	if overlaps(spanPoint2, spanPoint1) {
		return true
	}

	return false
}

func spansHasOverlap(s []Span) error {
	sort.Slice(s, func(i, j int) bool {
		return s[i].Start.Lt(s[j].Start)
	})

	if len(s) < 2 {
		return nil
	}

	for i := 0; i < len(s)-1; i++ {
		curr := s[i]
		next := s[i+1]
		if Overlaps(curr, next) {
			msg := fmt.Sprintf(
				"overlapping: (%s - %s) && (%s - %s)",
				curr.Start.Format("Y-m-d H:m"),
				curr.End.Format("Y-m-d H:m"),
				next.Start.Format("Y-m-d H:m"),
				next.End.Format("Y-m-d H:m"),
			)
			return errors.New(msg)
		}
	}

	return nil
}

// OverlappingSpans returns the overlapping time spans between two slices of time spans.
func OverlappingSpans(a, b []Span) ([]Span, error) {
	if err := spansHasOverlap(a); err != nil {
		return nil, errors.Join(ErrorSpansAHasOverlap, err)
	}
	if err := spansHasOverlap(b); err != nil {
		return nil, errors.Join(ErrorSpansBHasOverlap, err)
	}

	spOverlaps := make([]Span, 0)
	for _, spanA := range a {
		for _, spanB := range b {
			if Overlaps(spanA, spanB) {
				overlapSpan := OverlapSpan(spanA, spanB)
				if overlapSpan != nil {
					spOverlaps = append(spOverlaps, *overlapSpan)
				}
			}
		}
	}

	return spOverlaps, nil
}

// OverlapSpan returns the overlapping span between two time spans.
func OverlapSpan(a, b Span) *Span {
	if !Overlaps(a, b) {
		return nil
	}

	// we arrange all the points in time
	pointsInTime := []carbon.Carbon{
		a.Start,
		a.End,
		b.Start,
		b.End,
	}
	sort.Slice(pointsInTime, func(i, j int) bool {
		return pointsInTime[i].Lt(pointsInTime[j])
	})

	// and return the "between" points as the
	// overlapping spans (2nd, and 3rd points)
	return &Span{
		Start: pointsInTime[1],
		End:   pointsInTime[2],
	}
}
