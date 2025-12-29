package dateutil

import (
	"github.com/golang-module/carbon/v2"
	"time"
	"wingedapp/pgtester/internal/util/errutil"
)

const (
	layoutDate = "2006-01-02"
	layoutTS   = "2006-01-02T15:04:05.000Z"
)

func IsValid(s string) bool {
	_, err := time.Parse(layoutDate, s)
	return err != nil
}

func isValidStartAndEndDates(start, end string) error {
	var errs errutil.List

	t1 := carbon.Parse(start)
	t2 := carbon.Parse(end)

	if t1.IsInvalid() {
		errs.Add("start is invalid")
	}
	if t2.IsInvalid() {
		errs.Add("end is invalid")
	}
	if t1.Compare(">", t2) && (t1.IsValid() && t2.IsValid()) {
		errs.Add("start is greater than end")
	}

	return errs.Single()
}
