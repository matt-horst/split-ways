-- +goose Up
-- +goose StatementBegin
ALTER TABLE groups
ADD COLUMN owner UUID REFERENCES users(id) NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE groups
DROP COLUMN owner;
-- +goose StatementEnd
