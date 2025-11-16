-- +goose Up
-- +goose StatementBegin
ALTER TABLE expenses
ADD COLUMN amount NUMERIC(12, 2) NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE expenses
DROP COLUMN amount;
-- +goose StatementEnd
