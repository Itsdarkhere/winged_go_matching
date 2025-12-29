package errutil

import "strings"

type Wrapper struct {
	list List
}

func newWrapper(list List) *Wrapper {
	return &Wrapper{list}
}

func (w Wrapper) Error() string {
	return strings.Join(errsToArr(w.list), `", "`)
}

func (w Wrapper) ErrorArr() []string {
	return errsToArr(w.list)
}

func errsToArr(errs []error) []string {
	strs := make([]string, 0)
	for _, err := range errs {
		if err != nil {
			strs = append(strs, err.Error())
		}
	}
	return strs
}
