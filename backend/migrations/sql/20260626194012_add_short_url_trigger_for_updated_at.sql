-- +goose Up
DROP FUNCTION IF EXISTS set_updated_at();

-- +goose StatementBegin
CREATE FUNCTION set_updated_at()
RETURNS trigger AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER short_urls_set_updated_at
BEFORE UPDATE ON short_urls
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS short_urls_set_updated_at ON short_urls;
DROP FUNCTION IF EXISTS set_updated_at();
