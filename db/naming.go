package db

import "github.com/iancoleman/strcase"

// ToSnakeCase converts a string from camelCase or PascalCase to snake_case.
// This is used for automatic mapping of struct field names to database column names.
func ToSnakeCase(s string) string {
	return strcase.ToSnake(s)
}
