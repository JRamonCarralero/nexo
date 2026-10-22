-- Row-Level Security foundations (D-011).
--
-- The API never connects as the owner of the tables (owners bypass RLS). It
-- uses a login role that is a member of nexo_app, which only gets DML rights.
-- Login roles are provisioned per environment (see db/local/roles.sql for
-- local development and tests); this migration only needs nexo_app to exist.

-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'nexo_app') THEN
        CREATE ROLE nexo_app NOLOGIN;
    END IF;
END
$$;
-- +goose StatementEnd

GRANT USAGE ON SCHEMA public TO nexo_app;

-- Tables and sequences created from now on by the migration role.
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO nexo_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO nexo_app;

-- Request scope, published by the API in every transaction with
-- set_config('app.user_id' | 'app.workspace_id', value, true).
-- NULL when not set, so policies comparing with them match nothing.

-- +goose StatementBegin
CREATE FUNCTION nexo_user_id() RETURNS uuid
    LANGUAGE sql STABLE
AS $$ SELECT nullif(current_setting('app.user_id', true), '')::uuid $$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION nexo_workspace_id() RETURNS uuid
    LANGUAGE sql STABLE
AS $$ SELECT nullif(current_setting('app.workspace_id', true), '')::uuid $$;
-- +goose StatementEnd

-- +goose Down
DROP FUNCTION IF EXISTS nexo_workspace_id();
DROP FUNCTION IF EXISTS nexo_user_id();
ALTER DEFAULT PRIVILEGES IN SCHEMA public REVOKE USAGE, SELECT ON SEQUENCES FROM nexo_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public REVOKE SELECT, INSERT, UPDATE, DELETE ON TABLES FROM nexo_app;
REVOKE USAGE ON SCHEMA public FROM nexo_app;
-- The role itself is cluster-wide and may be shared: it is not dropped.
