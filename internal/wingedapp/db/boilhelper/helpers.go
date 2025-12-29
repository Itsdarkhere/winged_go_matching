package boilhelper

import (
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/business/sdk"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

// QmHelper is a helper for building query modifiers.
type QmHelper struct {
	tableName string
}

type QmColSet struct {
	TableName  string
	TableAlias string
	Cols       []QmCol
}

func QmGroupBy(tbl, col string) qm.QueryMod {
	return qm.GroupBy(fmt.Sprintf(
		"%s.%s",
		tbl,
		col,
	))
}

// QmJoinStr builds a join condition string with optional where clauses.
func QmJoinStr(extTbl, extTblCol, selfTbl, selfTblCol string, wheres ...string) string {
	condition := fmt.Sprintf(
		"%s %s ON %s.%s = %s.%s",
		extTbl,
		extTbl,
		extTbl,
		extTblCol,
		selfTbl,
		selfTblCol,
	)

	for _, where := range wheres {
		condition += fmt.Sprintf("AND %s", where)
	}

	return condition
}

// QmSelect convert a []QmColSet into an array select conditions.
func QmSelect(colSets []QmColSet) []string {
	allCols := make([]string, 0)
	for _, colSet := range colSets {
		cs := &QmHelper{colSet.TableName}
		cols := cs.Columns(colSet.Cols)
		allCols = append(allCols, cols...)
	}
	return allCols
}

type QmCol struct {
	Name     string
	Modifier string
	Alias    string
}

// Columns generates the column selection strings for the given Columns.
func (c *QmHelper) Columns(cols []QmCol) []string {
	selectCols := make([]string, 0)
	for _, col := range cols {
		// basic table.col
		selectedCol := fmt.Sprintf(
			"%s.%s",
			c.tableName,
			col.Name,
		)

		// add Modifier
		if col.Modifier != "" {
			selectedCol = fmt.Sprintf(col.Modifier, selectedCol)
		}

		// add (possible aliased) final column Name
		colName := col.Name
		if col.Alias != "" {
			colName = col.Alias
		}
		selectedCol = fmt.Sprintf("%s as %s", selectedCol, colName)

		selectCols = append(selectCols, selectedCol)
	}
	return selectCols
}

func ApplyPagination(qMods []qm.QueryMod, p *sdk.Pagination) []qm.QueryMod {
	if !p.Validate() {
		return qMods
	}

	if p.Rows.Valid {
		qMods = append(qMods, qm.Limit(p.Rows.Int))
	}

	if p.Page.Int > 1 {
		qMods = append(qMods, qm.Offset((p.Page.Int-1)*p.Rows.Int))
	}

	return qMods
}

// SortByAscOrDesc returns the SQL sort order based on the null.String input.
func SortByAscOrDesc(f null.String) string {
	s := "DESC"
	if f.Valid && f.String == "+" {
		s = "ASC"
	}
	return s
}
