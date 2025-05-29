-- +goose Up
-- +goose StatementBegin
BEGIN;

CREATE TABLE devices (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_uuid UUID NOT NULL REFERENCES users(uuid),
    name VARCHAR(50) NOT NULL,
    id CHAR(64) NOT NULL,
    master_key BYTEA,
    public_key BYTEA,
    confirmed BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_devices_user_uuid_name ON devices(user_uuid, name);

COMMIT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS devices;
-- +goose StatementEnd
