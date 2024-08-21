-- +goose Up
create table counter (
    name text primary key check(name ~ '^[^\.]*$'),
    labels text[]
);

create table counter_value (
    id text primary key,
    counter_name text references counter(name),
    label_values text[],
    value bigint check(value >= 0)
);

-- +goose Down
drop table counter_value cascade;
drop table counter cascade;
