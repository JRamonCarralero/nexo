-- +goose Up
-- citext: case-insensitive text (emails, slugs).
-- unaccent + pg_trgm: accent-insensitive and fuzzy search.
CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS unaccent;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- +goose Down
DROP EXTENSION IF EXISTS pg_trgm;
DROP EXTENSION IF EXISTS unaccent;
DROP EXTENSION IF EXISTS citext;
