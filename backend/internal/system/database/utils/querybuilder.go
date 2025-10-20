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
		// Handle nested JSON paths for PostgreSQL
		postgresJSONPath := buildPostgresJSONPath(columnName, key)
		postgresQuery += fmt.Sprintf(" AND %s = $%d", postgresJSONPath, i+1)
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

// buildPostgresJSONPath constructs the appropriate PostgreSQL JSON path expression.
// For nested paths (e.g., "address.city"), it uses the #>> operator with an array path.
// For single-level paths (e.g., "username"), it uses the ->> operator.
func buildPostgresJSONPath(columnName, key string) string {
	if !containsDot(key) {
		// Single-level path: use ->> operator
		return fmt.Sprintf("%s->>'%s'", columnName, key)
	}

	// Nested path: use #>> operator with array notation
	// Convert "address.city" to {address,city}
	pathParts := splitPath(key)
	pathArray := "{" + pathParts + "}"
	return fmt.Sprintf("%s#>>'%s'", columnName, pathArray)
}

// containsDot checks if a string contains a dot character.
func containsDot(s string) bool {
	for _, char := range s {
		if char == '.' {
			return true
		}
	}
	return false
}

// splitPath splits a dot-separated path into comma-separated parts.
// Example: "address.city" -> "address,city"
func splitPath(path string) string {
	parts := make([]string, 0)
	currentPart := ""

	for _, char := range path {
		if char == '.' {
			if currentPart != "" {
				parts = append(parts, currentPart)
				currentPart = ""
			}
		} else {
			currentPart += string(char)
		}
	}

	if currentPart != "" {
		parts = append(parts, currentPart)
	}

	result := ""
	for i, part := range parts {
		if i > 0 {
			result += ","
		}
		result += part
	}

	return result
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
