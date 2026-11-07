-- Invitations (D-011, PR 4; ADR-0024).
--
-- Two kinds:
--   email  for one person: single use, expires in 7 days. Accepting requires
--          an account with the same email, verified.
--   link   shareable: reusable until it expires (30 days) or is revoked.
--          It can never grant the admin role.
--
-- Only the SHA-256 hash of the token is stored: a database leak does not
-- expose usable invitations.
--
-- Admins of the current workspace see and revoke its invitations (RLS).
-- Creating and accepting go through SECURITY DEFINER functions: creating
-- needs to know whether an email already belongs to a member (users the admin
-- cannot see), and the person accepting is not a member yet, so RLS would not
-- let them create their membership.
--
-- Errors are raised as 'nexo:<code>' messages, mapped in the repository.

-- +goose Up
CREATE TABLE invitations (
    id           uuid PRIMARY KEY,
    workspace_id uuid NOT NULL REFERENCES workspaces (id),
    kind         text NOT NULL CHECK (kind IN ('email', 'link')),
    email        citext,
    role         text NOT NULL CHECK (role IN ('admin', 'member', 'guest')),
    token_hash   bytea NOT NULL UNIQUE,
    expires_at   timestamptz NOT NULL,
    accepted_at  timestamptz,
    accepted_by  uuid REFERENCES users (id),
    revoked_at   timestamptz,
    created_by   uuid REFERENCES users (id),
    created_at   timestamptz NOT NULL,
    updated_at   timestamptz NOT NULL,
    CHECK ((kind = 'email') = (email IS NOT NULL)),
    CHECK (kind = 'email' OR role <> 'admin')
);
CREATE INDEX invitations_pending_idx ON invitations (workspace_id)
    WHERE accepted_at IS NULL AND revoked_at IS NULL;

ALTER TABLE invitations ENABLE ROW LEVEL SECURITY;
CREATE POLICY invitations_select_admin ON invitations FOR SELECT
    USING (workspace_id = nexo_workspace_id() AND nexo_workspace_role(workspace_id) = 'admin');
CREATE POLICY invitations_update_admin ON invitations FOR UPDATE
    USING (workspace_id = nexo_workspace_id() AND nexo_workspace_role(workspace_id) = 'admin')
    WITH CHECK (workspace_id = nexo_workspace_id());

REVOKE INSERT, UPDATE, DELETE ON invitations FROM nexo_app;
GRANT UPDATE (revoked_at, updated_at) ON invitations TO nexo_app;

-- Creates an invitation in the current workspace. Admins only. Inviting an
-- email again revokes its previous pending invitation (resend).
-- +goose StatementBegin
CREATE FUNCTION nexo_create_invitation(
    p_id uuid, p_kind text, p_email citext, p_role text, p_token_hash bytea, p_expires_at timestamptz
) RETURNS void
    LANGUAGE plpgsql
    SECURITY DEFINER
    SET search_path = public, pg_temp
AS $$
DECLARE
    v_user uuid := nexo_user_id();
    v_ws   uuid := nexo_workspace_id();
BEGIN
    IF v_user IS NULL OR v_ws IS NULL OR nexo_workspace_role(v_ws) IS DISTINCT FROM 'admin' THEN
        RAISE EXCEPTION 'nexo:not_admin' USING ERRCODE = 'insufficient_privilege';
    END IF;
    IF p_kind = 'link' AND p_role = 'admin' THEN
        RAISE EXCEPTION 'nexo:link_cannot_grant_admin' USING ERRCODE = 'check_violation';
    END IF;

    IF p_kind = 'email' THEN
        IF EXISTS (
            SELECT 1 FROM memberships m JOIN users u ON u.id = m.user_id
             WHERE m.workspace_id = v_ws AND u.email = p_email
        ) THEN
            RAISE EXCEPTION 'nexo:already_member' USING ERRCODE = 'unique_violation';
        END IF;
        UPDATE invitations
           SET revoked_at = now(), updated_at = now()
         WHERE workspace_id = v_ws AND kind = 'email' AND email = p_email
           AND accepted_at IS NULL AND revoked_at IS NULL;
    END IF;

    INSERT INTO invitations (id, workspace_id, kind, email, role, token_hash, expires_at,
                             created_by, created_at, updated_at)
    VALUES (p_id, v_ws, p_kind, p_email, p_role, p_token_hash, p_expires_at, v_user, now(), now());
END
$$;
-- +goose StatementEnd

-- What the holder of a token is invited to, for the confirmation screen.
-- Returns no row for unknown tokens. Status: pending, expired, revoked or
-- accepted (email invitations only).
-- +goose StatementBegin
CREATE FUNCTION nexo_preview_invitation(p_token_hash bytea)
    RETURNS TABLE (workspace_id uuid, workspace_name text, kind text, email citext, role text,
                   invited_by_name text, expires_at timestamptz, status text)
    LANGUAGE sql STABLE
    SECURITY DEFINER
    SET search_path = public, pg_temp
AS $$
    SELECT w.id, w.name, i.kind, i.email, i.role, coalesce(u.name, ''), i.expires_at,
           CASE
               WHEN i.revoked_at IS NOT NULL THEN 'revoked'
               WHEN i.accepted_at IS NOT NULL THEN 'accepted'
               WHEN i.expires_at <= now() THEN 'expired'
               ELSE 'pending'
           END
      FROM invitations i
      JOIN workspaces w ON w.id = i.workspace_id AND w.deleted_at IS NULL
      LEFT JOIN users u ON u.id = i.created_by
     WHERE i.token_hash = p_token_hash
$$;
-- +goose StatementEnd

-- Accepts an invitation for the acting user and returns the workspace and the
-- role they have in it. Idempotent: someone already in the workspace keeps
-- their membership and role.
-- +goose StatementBegin
CREATE FUNCTION nexo_accept_invitation(p_token_hash bytea, p_membership_id uuid)
    RETURNS TABLE (workspace_id uuid, role text)
    LANGUAGE plpgsql
    SECURITY DEFINER
    SET search_path = public, pg_temp
AS $$
DECLARE
    v_user     uuid := nexo_user_id();
    v_inv      invitations%ROWTYPE;
    v_email    citext;
    v_verified boolean;
    v_role     text;
BEGIN
    IF v_user IS NULL THEN
        RAISE EXCEPTION 'nexo:no_user' USING ERRCODE = 'insufficient_privilege';
    END IF;

    SELECT i.* INTO v_inv
      FROM invitations i
      JOIN workspaces w ON w.id = i.workspace_id AND w.deleted_at IS NULL
     WHERE i.token_hash = p_token_hash
       FOR UPDATE OF i;
    IF NOT FOUND OR v_inv.revoked_at IS NOT NULL THEN
        RAISE EXCEPTION 'nexo:invitation_not_found' USING ERRCODE = 'no_data_found';
    END IF;

    SELECT m.role INTO v_role FROM memberships m
     WHERE m.workspace_id = v_inv.workspace_id AND m.user_id = v_user;

    IF v_inv.kind = 'email' THEN
        IF v_inv.accepted_at IS NOT NULL THEN
            IF v_inv.accepted_by = v_user AND v_role IS NOT NULL THEN
                RETURN QUERY SELECT v_inv.workspace_id, v_role;
                RETURN;
            END IF;
            RAISE EXCEPTION 'nexo:invitation_not_found' USING ERRCODE = 'no_data_found';
        END IF;
        IF v_inv.expires_at <= now() THEN
            RAISE EXCEPTION 'nexo:invitation_expired' USING ERRCODE = 'check_violation';
        END IF;
        SELECT u.email, u.email_verified INTO v_email, v_verified FROM users u WHERE u.id = v_user;
        IF v_email IS DISTINCT FROM v_inv.email THEN
            RAISE EXCEPTION 'nexo:invitation_email_mismatch' USING ERRCODE = 'insufficient_privilege';
        END IF;
        IF NOT v_verified THEN
            RAISE EXCEPTION 'nexo:email_not_verified' USING ERRCODE = 'insufficient_privilege';
        END IF;
    ELSIF v_inv.expires_at <= now() THEN
        RAISE EXCEPTION 'nexo:invitation_expired' USING ERRCODE = 'check_violation';
    END IF;

    IF v_role IS NULL THEN
        INSERT INTO memberships (id, workspace_id, user_id, role, created_by, created_at, updated_at)
        VALUES (p_membership_id, v_inv.workspace_id, v_user, v_inv.role, v_inv.created_by, now(), now());
        v_role := v_inv.role;
    END IF;

    IF v_inv.kind = 'email' THEN
        UPDATE invitations
           SET accepted_at = now(), accepted_by = v_user, updated_at = now()
         WHERE id = v_inv.id;
    END IF;

    RETURN QUERY SELECT v_inv.workspace_id, v_role;
END
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION nexo_create_invitation(uuid, text, citext, text, bytea, timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION nexo_preview_invitation(bytea) FROM PUBLIC;
REVOKE ALL ON FUNCTION nexo_accept_invitation(bytea, uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION nexo_create_invitation(uuid, text, citext, text, bytea, timestamptz) TO nexo_app;
GRANT EXECUTE ON FUNCTION nexo_preview_invitation(bytea) TO nexo_app;
GRANT EXECUTE ON FUNCTION nexo_accept_invitation(bytea, uuid) TO nexo_app;

-- +goose Down
DROP FUNCTION IF EXISTS nexo_accept_invitation(bytea, uuid);
DROP FUNCTION IF EXISTS nexo_preview_invitation(bytea);
DROP FUNCTION IF EXISTS nexo_create_invitation(uuid, text, citext, text, bytea, timestamptz);
DROP TABLE IF EXISTS invitations;
