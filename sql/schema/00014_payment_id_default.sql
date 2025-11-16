-- +goose Up
-- +goose StatementBegin
ALTER TABLE payments
ALTER COLUMN id SET DEFAULT GEN_RANDOM_UUID();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE payments
ALTER COLUMN id DROP DEFAULT;
-- +goose StatementEnd
