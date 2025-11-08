-- +goose Up
-- +goose StatementBegin
CREATE TABLE expenses (
    id UUID PRIMARY KEY,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    created_at TIME NOT NULL,
    updated_at TIME NOT NULL,
    paid_by UUID REFERENCES users(id) ON DELETE SET NULL,
    description TEXT NOT NULL,
    amount NUMERIC(12, 2) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE expenses;
-- +goose StatementEnd
