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
	unindexedFilters map[string]interface{},
	indexedFilters map[string]interface{},
) (model.DBQuery, []interface{}, error) {
	// Validate the column name.
	if err := validateKey(columnName); err != nil {
		return model.DBQuery{}, nil, fmt.Errorf("invalid column name: %w", err)
	}

	query := baseQuery

	args := make([]interface{}, 0, len(unindexedFilters) + len(indexedFilters))

	for columnName, value := range indexedFilters {
		query += fmt.Sprintf(" AND '%s' = $%d", columnName, len(args) + 1)
		args = append(args, value)
	}

	postgresQuery := query
	sqliteQuery := query

	unindexedKeys := make([]string, 0, len(unindexedFilters))
	for key := range unindexedFilters {
		if err := validateKey(key); err != nil {
			return model.DBQuery{}, nil, fmt.Errorf("invalid filter key: %w", err)
		}
		unindexedKeys = append(unindexedKeys, key)
	}
	sort.Strings(unindexedKeys)

	for _, key := range unindexedKeys {
		postgresQuery += fmt.Sprintf(" AND %s->>'%s' = $%d", columnName, key, len(args)+1)
		sqliteQuery += fmt.Sprintf(" AND json_extract(%s, '$.%s') = ?", columnName, key)
		args = append(args, unindexedFilters[key])
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
