package repository

import (
	"fmt"
	"strings"
)

type QueryBuilder struct {
	table      string
	selectCols []string
	whereCond  []string
	orderBy    []string
	groupBy    []string
	limit      int
	offset     int
	args       []interface{}
	argCounter int
}

func NewQueryBuilder(table string) *QueryBuilder {
	return &QueryBuilder{
		table:      table,
		selectCols: []string{"*"},
		argCounter: 1,
	}
}

func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.selectCols = columns
	return qb
}

func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	qb.whereCond = append(qb.whereCond, condition)
	qb.args = append(qb.args, args...)
	return qb
}

func (qb *QueryBuilder) WhereIf(condition string, apply bool, args ...interface{}) *QueryBuilder {
	if apply {
		return qb.Where(condition, args...)
	}
	return qb
}

func (qb *QueryBuilder) OrderBy(field string, direction string) *QueryBuilder {
	qb.orderBy = append(qb.orderBy, fmt.Sprintf("%s %s", field, direction))
	return qb
}

func (qb *QueryBuilder) GroupBy(fields ...string) *QueryBuilder {
	qb.groupBy = fields
	return qb
}

func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}

func (qb *QueryBuilder) Build() (string, []interface{}) {
	var query strings.Builder

	query.WriteString("SELECT ")
	query.WriteString(strings.Join(qb.selectCols, ", "))

	query.WriteString(" FROM ")
	query.WriteString(qb.table)

	if len(qb.whereCond) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(qb.whereCond, " AND "))
	}

	if len(qb.groupBy) > 0 {
		query.WriteString(" GROUP BY ")
		query.WriteString(strings.Join(qb.groupBy, ", "))
	}

	if len(qb.orderBy) > 0 {
		query.WriteString(" ORDER BY ")
		query.WriteString(strings.Join(qb.orderBy, ", "))
	}

	if qb.limit > 0 {
		query.WriteString(fmt.Sprintf(" LIMIT %d", qb.limit))
	}

	if qb.offset > 0 {
		query.WriteString(fmt.Sprintf(" OFFSET %d", qb.offset))
	}

	return query.String(), qb.args
}