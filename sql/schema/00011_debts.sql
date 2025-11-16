-- +goose Up
-- +goose StatementBegin
CREATE TABLE debts (
    id UUID PRIMARY KEY DEFAULT GEN_RANDOM_UUID(),
    expense_id UUID NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    owed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    owed_to UUID REFERENCES users(id) ON DELETE SET NULL,
    amount NUMERIC(12, 2) NOT NULL,
    CONSTRAINT owed_by_owed_to_not_equal CHECK (owed_by != owed_to)
);

DROP TABLE splits;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TABLE splits (
    id UUID PRIMARY KEY DEFAULT GEN_RANDOM_UUID(),
    expense_id UUID NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    amount NUMERIC(12, 2) NOT NULL
);


DROP TABLE debts;
-- +goose StatementEnd
