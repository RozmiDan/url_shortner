-- +goose Up
CREATE TABLE IF NOT EXISTS url(
    id SERIAL PRIMARY KEY,
    alias TEXT NOT NULL UNIQUE,
    url TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT now() NOT NULL
);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_url_timestamp() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW; -- исправлено здесь (была опечатка RETRUN)
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE OR REPLACE TRIGGER trigger_update_url_timestamp
    BEFORE UPDATE
    ON url
    FOR EACH ROW
EXECUTE FUNCTION update_url_timestamp();

-- +goose Down
DROP TRIGGER IF EXISTS trigger_update_url_timestamp ON url;
DROP FUNCTION IF EXISTS update_url_timestamp();
DROP TABLE IF EXISTS url;
