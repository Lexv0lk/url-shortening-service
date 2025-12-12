-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_url_token ON mappings (url_token);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_url_token;
-- +goose StatementEnd
