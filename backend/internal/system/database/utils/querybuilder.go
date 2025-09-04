// Package utils provides utility functions for database operations.
package utils

import (
	"fmt"
	"sort"

	"github.com/asgardeo/thunder/internal/system/database/model"
)

// BuildFilterQuery constructs a query to filter records based on the provided filters.
func BuildFilterQuery(
	queryID string,
	baseQuery string,
	columnName string,
	filters map[string]interface{},
) (model.DBQuery, []interface{}, error) {
	// Validate the column name.
	if err := validateKey(columnName); err != nil {
		return model.DBQuery{}, nil, fmt.Errorf("invalid column name: %w", err)
	}

	args := make([]interface{}, 0, len(filters))

	keys := make([]string, 0, len(filters))
	for key := range filters {
		if err := validateKey(key); err != nil {
			return model.DBQuery{}, nil, fmt.Errorf("invalid filter key: %w", err)
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	postgresQuery := baseQuery
	sqliteQuery := baseQuery
	for i, key := range keys {
		postgresQuery += fmt.Sprintf(" AND %s->>'%s' = $%d", columnName, key, i+1)
		sqliteQuery += fmt.Sprintf(" AND json_extract(%s, '$.%s') = ?", columnName, key)
		args = append(args, filters[key])
	}

	resultQuery := model.DBQuery{
		ID:            queryID,
		Query:         postgresQuery,
		PostgresQuery: postgresQuery,
		SQLiteQuery:   sqliteQuery,
	}

	return resultQuery, args, nil
}

// validateKey ensures that the provided key contains only safe characters (alphanumeric and underscores).
func validateKey(key string) error {
	for _, char := range key {
		if !(char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' ||
			char >= '0' && char <= '9' || char == '_' || char == '.') {
			return fmt.Errorf("key '%s' contains invalid characters", key)
		}
	}
	return nil
}
