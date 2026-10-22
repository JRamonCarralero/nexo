-- Users (D-011, PR 2).
--
-- Identity comes from the identity provider (Firebase): auth_subject is its
-- user id. A row may exist before the person first signs in (auth_subject
-- NULL), e.g. the bootstrap superadmin; it is linked on the first login with a
-- verified email.
--
-- The API role can only read and edit its own profile. Creating users and
-- linking identities happen exclusively through the two SECURITY DEFINER
-- functions below: narrow, audited entry points that run as the table owner.

-- +goose Up
CREATE TABLE users (
    id             uuid PRIMARY KEY,
    auth_subject   text UNIQUE,
    email          citext NOT NULL UNIQUE,
    email_verified boolean NOT NULL DEFAULT false,
    name           text NOT NULL DEFAULT '',
    avatar_url     text NOT NULL DEFAULT '',
    is_superadmin  boolean NOT NULL DEFAULT false,
    created_at     timestamptz NOT NULL,
    updated_at     timestamptz NOT NULL
);

ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE users FORCE ROW LEVEL SECURITY;

-- A user sees and edits their own row. (Workspaces will add: people you share a workspace with.)
CREATE POLICY users_select_self ON users FOR SELECT
    USING (id = nexo_user_id());
CREATE POLICY users_update_self ON users FOR UPDATE
    USING (id = nexo_user_id())
    WITH CHECK (id = nexo_user_id());

-- RLS limits rows, not columns: without this a user could make themselves
-- superadmin. Only profile columns are editable; inserts and deletes go
-- through the functions below.
REVOKE INSERT, UPDATE, DELETE ON users FROM nexo_app;
GRANT UPDATE (name, avatar_url, updated_at) ON users TO nexo_app;

-- nexo_resolve_user maps a verified identity to a Nexo user, creating it on
-- the first login. Returns the user id.
--   1. Known identity → refresh email data, fill empty profile fields.
--   2. Pre-provisioned row with the same email and no identity yet → link it,
--      only when the provider says the email is verified.
--   3. Otherwise → create the user (p_new_id comes from the API: UUIDv7).
-- An email already used by another account raises a unique violation.

-- +goose StatementBegin
CREATE FUNCTION nexo_resolve_user(
    p_new_id uuid, p_subject text, p_email citext, p_email_verified boolean,
    p_name text, p_avatar_url text
) RETURNS uuid
    LANGUAGE plpgsql
    SECURITY DEFINER
    SET search_path = public, pg_temp
AS $$
DECLARE
    v_id uuid;
BEGIN
    UPDATE users
       SET email          = p_email,
           email_verified = p_email_verified,
           name           = CASE WHEN name = '' THEN p_name ELSE name END,
           avatar_url     = CASE WHEN avatar_url = '' THEN p_avatar_url ELSE avatar_url END,
           updated_at     = now()
     WHERE auth_subject = p_subject
    RETURNING id INTO v_id;
    IF FOUND THEN
        RETURN v_id;
    END IF;

    IF p_email_verified THEN
        UPDATE users
           SET auth_subject   = p_subject,
               email_verified = true,
               name           = CASE WHEN name = '' THEN p_name ELSE name END,
               avatar_url     = CASE WHEN avatar_url = '' THEN p_avatar_url ELSE avatar_url END,
               updated_at     = now()
         WHERE email = p_email AND auth_subject IS NULL
        RETURNING id INTO v_id;
        IF FOUND THEN
            RETURN v_id;
        END IF;
    END IF;

    INSERT INTO users (id, auth_subject, email, email_verified, name, avatar_url, created_at, updated_at)
    VALUES (p_new_id, p_subject, p_email, p_email_verified, p_name, p_avatar_url, now(), now());
    RETURN p_new_id;
END
$$;
-- +goose StatementEnd

-- nexo_bootstrap_superadmin makes sure the platform has a superadmin. It never
-- demotes anyone and does nothing if a superadmin already exists. Returns:
--   'exists'     a superadmin already exists
--   'promoted'   an existing user with that email became superadmin
--   'created'    a pre-provisioned row was created (linked on first login)
--   'unverified' a user with that email exists but has not verified it: not
--                promoted, as anyone can register an unverified email

-- +goose StatementBegin
CREATE FUNCTION nexo_bootstrap_superadmin(p_new_id uuid, p_email citext) RETURNS text
    LANGUAGE plpgsql
    SECURITY DEFINER
    SET search_path = public, pg_temp
AS $$
BEGIN
    IF EXISTS (SELECT 1 FROM users WHERE is_superadmin) THEN
        RETURN 'exists';
    END IF;

    UPDATE users
       SET is_superadmin = true, updated_at = now()
     WHERE email = p_email AND (email_verified OR auth_subject IS NULL);
    IF FOUND THEN
        RETURN 'promoted';
    END IF;

    IF EXISTS (SELECT 1 FROM users WHERE email = p_email) THEN
        RETURN 'unverified';
    END IF;

    INSERT INTO users (id, email, is_superadmin, created_at, updated_at)
    VALUES (p_new_id, p_email, true, now(), now());
    RETURN 'created';
END
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION nexo_resolve_user(uuid, text, citext, boolean, text, text) FROM PUBLIC;
REVOKE ALL ON FUNCTION nexo_bootstrap_superadmin(uuid, citext) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION nexo_resolve_user(uuid, text, citext, boolean, text, text) TO nexo_app;
GRANT EXECUTE ON FUNCTION nexo_bootstrap_superadmin(uuid, citext) TO nexo_app;

-- +goose Down
DROP FUNCTION IF EXISTS nexo_bootstrap_superadmin(uuid, citext);
DROP FUNCTION IF EXISTS nexo_resolve_user(uuid, text, citext, boolean, text, text);
DROP TABLE IF EXISTS users;
