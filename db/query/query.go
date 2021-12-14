package query

import "github.com/Masterminds/squirrel"

type Query struct {
	squirrel.StatementBuilderType
}

func New() Query {
	return Query{squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}
}
