-- +goose Up
-- +goose StatementBegin
create table photos
(
    path                text not null unique,
    longitude           text not null,
    latitude            text not null,
    country             text,
    city                text,
    original_created_at text not null,
    created_at          text not null,
    data                text not null
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table photos;
-- +goose StatementEnd
