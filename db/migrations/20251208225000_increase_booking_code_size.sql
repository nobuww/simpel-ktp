-- +goose Up
-- +goose StatementBegin
ALTER TABLE permohonan ALTER COLUMN kode_booking TYPE VARCHAR(20);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE permohonan ALTER COLUMN kode_booking TYPE CHAR(10);
-- +goose StatementEnd
