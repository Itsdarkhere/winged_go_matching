package timerange

import "errors"

var (
	ErrorSpansAHasOverlap = errors.New("spans in slice A has overlap")
	ErrorSpansBHasOverlap = errors.New("spans in slice B has overlap")
)
