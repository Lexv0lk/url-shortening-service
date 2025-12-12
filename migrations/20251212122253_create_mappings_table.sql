-- +goose Up
-- +goose StatementBegin
CREATE TABLE mappings (
    id           BIGSERIAL PRIMARY KEY,
    original_url TEXT NOT NULL,
    url_token    TEXT NOT NULL UNIQUE,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE mappings;
-- +goose StatementEnd
