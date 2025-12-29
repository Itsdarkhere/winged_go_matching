package sdk

import (
	"github.com/aarondl/null/v8"
)

// Pagination is an updated version of Pagination with potential new fields or methods.
type Pagination struct {
	Rows  null.Int `json:"rows"`
	Page  null.Int `json:"page"`
	Total int      `json:"total_rows"`
	Pages []int    `json:"pages"`
}

func (p *Pagination) Validate() bool {
	return p != nil
}

func NewDefaultPagination() *Pagination {
	return &Pagination{
		Rows: null.IntFrom(defaultRows),
		Page: null.IntFrom(defaultPage),
	}
}

// Recalculated recalculates the pagination based on the total entries.
func (p *Pagination) Recalculated(totalEntries int) *Pagination {
	p2 := p
	if p == nil {
		p2 = NewDefaultPagination()
	}

	// set total from items
	p2.Total = totalEntries

	// recalculate rows, and page
	if p2.Rows.Int > p2.Total {
		p2.Rows = null.IntFrom(p2.Total)
		p2.Page = null.IntFrom(1)
	}

	// recalculate pages
	p2.Pages = make([]int, 0)
	pageAdded := 1
	decrement := p2.Total
	for decrement > 0 {
		p2.Pages = append(p2.Pages, pageAdded)
		pageAdded++
		decrement -= p2.Rows.Int
	}

	return p2
}

// Recalculate recalculates the pagination based on the total entries.
func (p *Pagination) Recalculate(totalEntries int) {
	if p == nil {
		return // guard
	}

	// set total from items
	p.Total = totalEntries

	// recalculate rows, and page
	if p.Rows.Int > p.Total {
		p.Rows = null.IntFrom(p.Total)
		p.Page = null.IntFrom(1)
	}

	// recalculate pages
	p.Pages = make([]int, 0)
	pageAdded := 1
	decrement := p.Total
	for decrement > 0 {
		p.Pages = append(p.Pages, pageAdded)
		pageAdded++
		decrement -= p.Rows.Int
	}
}
