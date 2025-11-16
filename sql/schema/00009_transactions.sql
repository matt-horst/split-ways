-- +goose Up
-- +goose StatementBegin
CREATE TABLE transactions (
    id UUID PRIMARY KEY,
    created_at TIME NOT NULL DEFAULT NOW(),
    updated_at TIME NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    group_id UUID REFERENCES groups(id) ON DELETE CASCADE
);

ALTER TABLE expenses 
ADD COLUMN transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
DROP COLUMN created_at,
DROP COLUMN updated_at,
DROP COLUMN created_by,
DROP COLUMN group_id;

ALTER TABLE payments
ADD COLUMN transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
DROP COLUMN created_at,
DROP COLUMN updated_at,
DROP COLUMN created_by,
DROP COLUMN group_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE payments
DROP COLUMN transaction_id,
ADD COLUMN created_at TIME NOT NULL,
ADD COLUMN updated_at TIME NOT NULL,
ADD COLUMN created_by UUID REFERENCES users(id) ON DELETE SET NULL,
ADD COLUMN group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE;

ALTER TABLE expenses
DROP COLUMN transaction_id,
ADD COLUMN created_at TIME NOT NULL,
ADD COLUMN updated_at TIME NOT NULL,
ADD COLUMN created_by UUID REFERENCES users(id) ON DELETE SET NULL,
ADD COLUMN group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE;


DROP TABLE transactions;
-- +goose StatementEnd
