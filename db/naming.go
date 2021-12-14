package db

import "github.com/iancoleman/strcase"

func ToSnakeCase(s string) string {
	return strcase.ToSnake(s)
}
