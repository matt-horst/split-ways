-- +goose Up
-- +goose StatementBegin
CREATE TABLE payments (
    id UUID PRIMARY KEY,
    created_at TIME NOT NULL,
    updated_at TIME NOT NULL,
    group_id UUID REFERENCES groups(id) ON DELETE CASCADE,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    paid_by UUID REFERENCES users(id) ON DELETE SET NULL,
    paid_to UUID REFERENCES users(id) ON DELETE SET NULL,
    amount NUMERIC(12, 2) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE payments;
-- +goose StatementEnd
