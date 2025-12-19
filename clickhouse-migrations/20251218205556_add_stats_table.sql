-- +goose Up
-- +goose StatementBegin
CREATE TABLE stats_events
(
    id          UUID DEFAULT generateUUIDv4(),
    url_token   String,
    timestamp   DateTime64(3, 'UTC'),
    country     String,
    city        String,
    device_type String,
    referrer    String
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(timestamp)
ORDER BY (url_token, timestamp, id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE stats_events;
-- +goose StatementEnd
