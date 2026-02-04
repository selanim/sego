package repo

import (
	"fmt"
	"strings"
)

// QueryBuilder builds SQL queries dynamically
type QueryBuilder struct {
	tableName   string
	selectCols  []string
	whereClause []string
	orderBy     string
	groupBy     string
	having      string
	limit       int
	offset      int
	args        []interface{}
	joinClauses []string
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(tableName string) *QueryBuilder {
	return &QueryBuilder{
		tableName:  tableName,
		selectCols: []string{"*"},
	}
}

// Select specifies columns to select
func (qb *QueryBuilder) Select(cols ...string) *QueryBuilder {
	if len(cols) > 0 {
		qb.selectCols = cols
	}
	return qb
}

// Where adds a WHERE condition
func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	qb.whereClause = append(qb.whereClause, condition)
	qb.args = append(qb.args, args...)
	return qb
}

// WhereEq adds an equality WHERE condition
func (qb *QueryBuilder) WhereEq(field string, value interface{}) *QueryBuilder {
	qb.whereClause = append(qb.whereClause, fmt.Sprintf("%s = $%d", field, len(qb.args)+1))
	qb.args = append(qb.args, value)
	return qb
}

// WhereIn adds an IN WHERE condition
func (qb *QueryBuilder) WhereIn(field string, values []interface{}) *QueryBuilder {
	if len(values) == 0 {
		return qb
	}

	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = fmt.Sprintf("$%d", len(qb.args)+i+1)
	}

	qb.whereClause = append(qb.whereClause, fmt.Sprintf("%s IN (%s)", field, strings.Join(placeholders, ", ")))
	qb.args = append(qb.args, values...)
	return qb
}

// WhereLike adds a LIKE WHERE condition
func (qb *QueryBuilder) WhereLike(field string, pattern string) *QueryBuilder {
	qb.whereClause = append(qb.whereClause, fmt.Sprintf("%s LIKE $%d", field, len(qb.args)+1))
	qb.args = append(qb.args, "%"+pattern+"%")
	return qb
}

// OrderBy adds ORDER BY clause
func (qb *QueryBuilder) OrderBy(field string, desc ...bool) *QueryBuilder {
	direction := "ASC"
	if len(desc) > 0 && desc[0] {
		direction = "DESC"
	}
	qb.orderBy = fmt.Sprintf("%s %s", field, direction)
	return qb
}

// GroupBy adds GROUP BY clause
func (qb *QueryBuilder) GroupBy(fields ...string) *QueryBuilder {
	if len(fields) > 0 {
		qb.groupBy = strings.Join(fields, ", ")
	}
	return qb
}

// Having adds HAVING clause
func (qb *QueryBuilder) Having(condition string, args ...interface{}) *QueryBuilder {
	qb.having = condition
	qb.args = append(qb.args, args...)
	return qb
}

// Limit adds LIMIT clause
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

// Offset adds OFFSET clause
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}

// Join adds a JOIN clause
func (qb *QueryBuilder) Join(joinType, table, onCondition string) *QueryBuilder {
	qb.joinClauses = append(qb.joinClauses, fmt.Sprintf("%s JOIN %s ON %s", joinType, table, onCondition))
	return qb
}

// InnerJoin adds an INNER JOIN clause
func (qb *QueryBuilder) InnerJoin(table, onCondition string) *QueryBuilder {
	return qb.Join("INNER", table, onCondition)
}

// LeftJoin adds a LEFT JOIN clause
func (qb *QueryBuilder) LeftJoin(table, onCondition string) *QueryBuilder {
	return qb.Join("LEFT", table, onCondition)
}

// RightJoin adds a RIGHT JOIN clause
func (qb *QueryBuilder) RightJoin(table, onCondition string) *QueryBuilder {
	return qb.Join("RIGHT", table, onCondition)
}

// Build builds the SQL query
func (qb *QueryBuilder) Build() (string, []interface{}) {
	var query strings.Builder

	// SELECT
	query.WriteString("SELECT ")
	query.WriteString(strings.Join(qb.selectCols, ", "))
	query.WriteString(" FROM ")
	query.WriteString(qb.tableName)

	// JOINs
	for _, join := range qb.joinClauses {
		query.WriteString(" ")
		query.WriteString(join)
	}

	// WHERE
	if len(qb.whereClause) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(qb.whereClause, " AND "))
	}

	// GROUP BY
	if qb.groupBy != "" {
		query.WriteString(" GROUP BY ")
		query.WriteString(qb.groupBy)
	}

	// HAVING
	if qb.having != "" {
		query.WriteString(" HAVING ")
		query.WriteString(qb.having)
	}

	// ORDER BY
	if qb.orderBy != "" {
		query.WriteString(" ORDER BY ")
		query.WriteString(qb.orderBy)
	}

	// LIMIT
	if qb.limit > 0 {
		query.WriteString(fmt.Sprintf(" LIMIT %d", qb.limit))
	}

	// OFFSET
	if qb.offset > 0 {
		query.WriteString(fmt.Sprintf(" OFFSET %d", qb.offset))
	}

	return query.String(), qb.args
}

// BuildCount builds a COUNT query
func (qb *QueryBuilder) BuildCount() (string, []interface{}) {
	var query strings.Builder

	query.WriteString("SELECT COUNT(*) FROM ")
	query.WriteString(qb.tableName)

	// JOINs
	for _, join := range qb.joinClauses {
		query.WriteString(" ")
		query.WriteString(join)
	}

	// WHERE
	if len(qb.whereClause) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(qb.whereClause, " AND "))
	}

	return query.String(), qb.args
}
