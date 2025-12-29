package skd

import (
	"fmt"
	"wingedapp/pgtester/internal/util/validationlib"

	"github.com/aarondl/null/v8"
)

const (
	defaultPaginationPage    = 1
	defaultPaginationMaxRows = 10
	maxPaginationRows        = 100
)

type Pagination struct {
	Pages      []int `json:"pages"`
	Page       int   `mapstructure:"page" validate:"is_positive" json:"page"`
	Offset     int   `json:"-"`
	RowCount   int   `json:"row_count"`
	TotalCount int   `json:"total_count"`
	MaxRows    int   `mapstructure:"max_rows" validate:"is_positive" json:"max_rows"`
}

// SetQueryBoundaries sets the pagination boundaries for a query, namely:
// Page, MaxRows, TotalCount
func (p *Pagination) SetQueryBoundaries(page, maxRows, totalCount int) {
	p.TotalCount = totalCount

	// Ensure the page is within allowed limits
	if page < 1 {
		p.Page = defaultPaginationPage
	} else {
		p.Page = page
	}

	if maxRows < 1 {
		p.MaxRows = defaultPaginationMaxRows
	} else {
		p.MaxRows = maxRows
	}

	totalPages := totalCount / p.MaxRows
	if totalCount == 0 {
		totalPages = 1
	} else {
		if totalCount%p.MaxRows > 0 {
			totalPages++
		}
	}

	// Populate the Pages slice
	p.Pages = make([]int, totalPages)
	for i := range p.Pages {
		p.Pages[i] = i + 1
	}

	// Adjust the Page number to be within the calculated total pages
	if p.Page > totalPages {
		p.Page = 1 // Reset to page 1 if out of bounds
	}

	p.Offset = (p.Page - 1) * p.MaxRows
}

func (p *PaginationQueryFilters) SetDefaults() {
	if p.Page.Valid {
		if p.Page.Int < 1 || p.Page.Int == 0 {
			p.Page.Int = defaultPaginationPage
		}
	} else {
		p.Page.Int = defaultPaginationPage
	}

	if p.Page.Valid {
		if p.MaxRows.Int < 1 || p.MaxRows.Int == 0 {
			p.MaxRows.Int = defaultPaginationMaxRows
		}
	} else {
		p.MaxRows.Int = defaultPaginationMaxRows
	}
}

func NewPagination() *Pagination {
	return &Pagination{
		MaxRows: defaultPaginationMaxRows,
		Page:    defaultPaginationPage,
	}
}

func (p *Pagination) ValidatePagination() error {
	err := validationlib.Validate(p)
	if err != nil {
		return fmt.Errorf("validate: %v", err)
	}

	if p.MaxRows > maxPaginationRows {
		return fmt.Errorf("max pagination rows exceeded: %v", p.MaxRows)
	}

	return err
}

func (p *PaginationQueryFilters) Validate() error {
	if err := validationlib.Validate(p); err != nil {
		return fmt.Errorf("validate: %v", err)
	}

	if p.MaxRows.Valid {
		if p.MaxRows.Int > maxPaginationRows {
			return fmt.Errorf("max pagination rows exceeded: %v", p.MaxRows)
		}
	}

	return nil
}

type PaginationQueryFilters struct {
	MaxRows null.Int `query:"max_rows" mapstructure:"max_rows" json:"max_rows"`
	Page    null.Int `query:"page" mapstructure:"page" json:"page"`
}
