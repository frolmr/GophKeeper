-- +goose Up
-- +goose StatementBegin
BEGIN;

CREATE TABLE records (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_uuid UUID NOT NULL REFERENCES users(uuid),
    name VARCHAR(100) NOT NULL,
    kind VARCHAR(20) NOT NULL,
    payload BYTEA NOT NULL,
    metadata BYTEA
);

CREATE INDEX idx_records_user_uuid ON records(user_uuid);

COMMIT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS records;
-- +goose StatementEnd
