-- +goose Up
create table if not exists message (
    id serial,
    title text unique,
    text text,
    checked boolean,
    hash text
);

create index if not exists idx_message on message(title);

-- +goose Down
drop table if exists message;
drop index if exists idx_message;
