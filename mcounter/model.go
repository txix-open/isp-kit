package mcounter

import "github.com/lib/pq"

type counter struct {
	Name   string         `db:"name"`
	Labels pq.StringArray `db:"labels"`

	counterValues map[string]*counterValue
}

type counterValue struct {
	// id = hash(CounterName + LabelValues)
	Id          string         `db:"id"`
	CounterName string         `db:"counter_name"`
	LabelValues pq.StringArray `db:"label_values"`
	AddValue    int            `db:"counter_value"`
}
