-- Row-Level Security is ENABLED but not FORCED (ADR-0022).
--
-- FORCE only makes the table owner subject to the policies. The API never
-- connects as the owner (it uses nexo_api), so FORCE protected nothing; but it
-- did block the owner's own sessions: SECURITY DEFINER functions such as
-- nexo_resolve_user run as the owner and must bypass the policies.
--
-- It went unnoticed because locally and in the first tests the owner was a
-- superuser, and superusers always bypass RLS. In managed databases (Neon,
-- Cloud SQL) the owner is not a superuser. Tests now run migrations as a
-- non-superuser owner.

-- +goose Up
ALTER TABLE users NO FORCE ROW LEVEL SECURITY;

-- +goose Down
ALTER TABLE users FORCE ROW LEVEL SECURITY;
