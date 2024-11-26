package dbutil

import (
	"fmt"
	"strings"
)

type Filter struct {
	Field string
	Value interface{}
	Op    string
}

type QueryBuilder struct {
	baseQuery string
	filters   []Filter
	limit     int
	offset    int
}

func NewQueryBuilder(baseQuery string) *QueryBuilder {
	return &QueryBuilder{baseQuery: baseQuery}
}

func (qb *QueryBuilder) AddFilter(field, op string, value interface{}) *QueryBuilder {
	qb.filters = append(qb.filters, Filter{Field: field, Op: op, Value: value})
	return qb
}

func (qb *QueryBuilder) SetPagination(limit, offset int) *QueryBuilder {
	qb.limit = limit
	qb.offset = offset
	return qb
}

func (qb *QueryBuilder) Build() (string, []interface{}) {
	query := qb.baseQuery
	var args []interface{}
	argIndex := 1

	if len(qb.filters) > 0 {
		var conditions []string
		for _, filter := range qb.filters {
			conditions = append(conditions, fmt.Sprintf("%s %s $%d", filter.Field, filter.Op, argIndex))
			args = append(args, filter.Value)
			argIndex++
		}
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	if qb.limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, qb.limit)
		argIndex++
	}

	if qb.offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, qb.offset)
	}

	return query, args
}
