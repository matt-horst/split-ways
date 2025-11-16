-- +goose Up
-- +goose StatementBegin
ALTER TABLE transactions
ALTER COLUMN group_id SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE transactions
ALTER COLUMN group_id DROP NOT NULL;
-- +goose StatementEnd
