package mcounter

import (
	"github.com/txix-open/isp-kit/db"
	"golang.org/x/net/context"
)

type CounterRepo struct {
	cli db.DB
}

func NewCounterRepo(cli db.DB) *CounterRepo {
	return &CounterRepo{
		cli: cli,
	}
}

func (c *CounterRepo) Counters(ctx context.Context) ([]counter, error) {
	var counters []counter
	err := c.cli.Select(ctx, &counters, "select * from counter")
	if err != nil {
		return nil, err
	}
	return counters, nil
}

func (c *CounterRepo) CounterValues(ctx context.Context, counterName string) ([]counterValue, error) {
	var counterValues []counterValue
	err := c.cli.Select(ctx, &counterValues, "select * from counter_value where counter_name = $1", counterName)
	if err != nil {
		return nil, err
	}
	return counterValues, nil
}

func (c *CounterRepo) UpsertCounter(ctx context.Context, counters []*counter) error {
	for _, counter := range counters {
		_, err := c.cli.Exec(ctx, `insert into counter values ($1, $2) 
                  on conflict (name) do update set labels = EXCLUDED.labels`,
			counter.Name, counter.Labels)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CounterRepo) UpsertCounterValue(ctx context.Context, counterValues []*counterValue) error {
	for _, counterVal := range counterValues {
		_, err := c.cli.Exec(ctx,
			`insert into counter_value values($1, $2, $3, $4) on conflict (id) 
			do update set label_values = EXCLUDED.label_values,
 			value = counter_value.value + EXCLUDED.value,
 			counter_name = EXCLUDED.counter_name`,
			counterVal.Id, counterVal.CounterName, counterVal.LabelValues, counterVal.AddValue)
		if err != nil {
			return err
		}
	}
	return nil
}
