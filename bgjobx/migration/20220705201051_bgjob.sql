-- +goose Up
create table bgjob_job
(
    id          text primary key,
    queue       text      not null,
    type        text      not null,
    arg         bytea,
    attempt     int4      not null,
    last_error  text,
    next_run_at int8      not null,
    created_at  timestamp not null,
    updated_at  timestamp not null
);

create index ix_bgjob_job__queue_next_run_at_created_at on bgjob_job (queue, next_run_at, created_at);

create table bgjob_dead_job
(
    id             serial8 primary key,
    job_id         text      not null,
    queue          text      not null,
    type           text      not null,
    arg            bytea,
    attempt        int4      not null,
    last_error     text      not null,
    next_run_at    int8      not null,
    job_created_at timestamp not null,
    job_updated_at timestamp not null,
    moved_at       timestamp not null
);

-- +goose Down
