-- +goose Up
-- +goose StatementBegin
CREATE TABLE stats_events (
    id              BIGSERIAL PRIMARY KEY,
    url_token       TEXT NOT NULL,
    timestamp       TIMESTAMP WITH TIME ZONE,
    country         TEXT,
    city            TEXT,
    device_type     TEXT,
    referrer        TEXT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE stats_events;
-- +goose StatementEnd
