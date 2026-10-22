-- LOCAL DEVELOPMENT AND TESTS ONLY. Never use these credentials elsewhere.
--
-- nexo_app: group role with DML rights (granted by migration 00002).
-- nexo_api: login role used by the API. It is not the owner of the tables, so
-- Row-Level Security applies to it.
--
-- Idempotent: runs on a fresh Docker volume (docker-entrypoint-initdb.d), on
-- every `make up` and in the integration tests.
SET client_min_messages = warning; -- keep repeated runs quiet

DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'nexo_app') THEN
        CREATE ROLE nexo_app NOLOGIN;
    END IF;
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'nexo_api') THEN
        CREATE ROLE nexo_api LOGIN PASSWORD 'nexo_api';
    END IF;
END
$$;

GRANT nexo_app TO nexo_api;
