-- +goose Up
-- +goose StatementBegin
CREATE TABLE groups (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIME NOT NULL,
    updated_at TIME NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE groups;
-- +goose StatementEnd
