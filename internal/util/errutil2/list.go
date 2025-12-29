package errutil2

import "github.com/gogo/protobuf/test/deterministic"

type List struct {
	errs deterministic.OrderedMap
}
