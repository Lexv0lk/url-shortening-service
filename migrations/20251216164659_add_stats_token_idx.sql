-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_stats_token ON stats_events (url_token);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_stats_token;
-- +goose StatementEnd
