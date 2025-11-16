-- +goose Up
-- +goose StatementBegin
ALTER TABLE transactions
ADD COLUMN kind TEXT NOT NULL DEFAULT 'unknown';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE transactions
DROP COLUMN kind;
-- +goose StatementEnd
