-- Workspaces and memberships (D-011, PR 3).
--
-- Visibility rules, enforced by Row-Level Security:
--   workspaces   a user sees the workspaces they belong to; admins of the
--                current workspace can rename it.
--   memberships  a user sees their own memberships (to list "my workspaces")
--                and every membership of the current workspace they belong to;
--                admins of the current workspace change roles and remove
--                members; anyone can remove themselves (leave).
--   users        widened: people who share a workspace see each other.
--
-- Rows are created only through SECURITY DEFINER functions (creating a
-- workspace adds its first admin atomically; invitations will add the rest).
-- A trigger guarantees that a workspace always keeps at least one admin.
--
-- Helper functions are SECURITY DEFINER so that policies on memberships can
-- query memberships without recursing into their own policies.

-- +goose Up
CREATE TABLE workspaces (
    id         uuid PRIMARY KEY,
    name       text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 100),
    created_by uuid REFERENCES users (id),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz
);

CREATE TABLE memberships (
    id           uuid PRIMARY KEY,
    workspace_id uuid NOT NULL REFERENCES workspaces (id),
    user_id      uuid NOT NULL REFERENCES users (id),
    role         text NOT NULL CHECK (role IN ('admin', 'member', 'guest')),
    created_by   uuid REFERENCES users (id),
    created_at   timestamptz NOT NULL,
    updated_at   timestamptz NOT NULL,
    UNIQUE (workspace_id, user_id)
);
CREATE INDEX memberships_user_idx ON memberships (user_id);

-- Role of the acting user in a workspace (NULL if not a member or the
-- workspace is deleted).
-- +goose StatementBegin
CREATE FUNCTION nexo_workspace_role(p_workspace uuid) RETURNS text
    LANGUAGE sql STABLE
    SECURITY DEFINER
    SET search_path = public, pg_temp
AS $$
    SELECT m.role
      FROM memberships m
      JOIN workspaces w ON w.id = m.workspace_id AND w.deleted_at IS NULL
     WHERE m.workspace_id = p_workspace AND m.user_id = nexo_user_id()
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION nexo_is_member(p_workspace uuid) RETURNS boolean
    LANGUAGE sql STABLE
    SECURITY DEFINER
    SET search_path = public, pg_temp
AS $$ SELECT nexo_workspace_role(p_workspace) IS NOT NULL $$;
-- +goose StatementEnd

-- Whether the acting user shares a (non-deleted) workspace with p_user.
-- +goose StatementBegin
CREATE FUNCTION nexo_shares_workspace(p_user uuid) RETURNS boolean
    LANGUAGE sql STABLE
    SECURITY DEFINER
    SET search_path = public, pg_temp
AS $$
    SELECT EXISTS (
        SELECT 1
          FROM memberships mine
          JOIN memberships theirs ON theirs.workspace_id = mine.workspace_id
          JOIN workspaces w ON w.id = mine.workspace_id AND w.deleted_at IS NULL
         WHERE mine.user_id = nexo_user_id() AND theirs.user_id = p_user
    )
$$;
-- +goose StatementEnd

-- Creates a workspace with the acting user as its first admin.
-- +goose StatementBegin
CREATE FUNCTION nexo_create_workspace(p_workspace_id uuid, p_membership_id uuid, p_name text) RETURNS void
    LANGUAGE plpgsql
    SECURITY DEFINER
    SET search_path = public, pg_temp
AS $$
DECLARE
    v_user uuid := nexo_user_id();
BEGIN
    IF v_user IS NULL THEN
        RAISE EXCEPTION 'nexo:no_user' USING ERRCODE = 'insufficient_privilege';
    END IF;
    INSERT INTO workspaces (id, name, created_by, created_at, updated_at)
    VALUES (p_workspace_id, p_name, v_user, now(), now());
    INSERT INTO memberships (id, workspace_id, user_id, role, created_by, created_at, updated_at)
    VALUES (p_membership_id, p_workspace_id, v_user, 'admin', v_user, now(), now());
END
$$;
-- +goose StatementEnd

-- A workspace must always keep at least one admin. The workspace row is
-- locked so that two concurrent demotions cannot both pass the check.
-- +goose StatementBegin
CREATE FUNCTION nexo_keep_one_admin() RETURNS trigger
    LANGUAGE plpgsql
    SECURITY DEFINER
    SET search_path = public, pg_temp
AS $$
BEGIN
    IF OLD.role = 'admin' AND (TG_OP = 'DELETE' OR NEW.role <> 'admin') THEN
        PERFORM 1 FROM workspaces WHERE id = OLD.workspace_id FOR UPDATE;
        IF NOT EXISTS (
            SELECT 1 FROM memberships
             WHERE workspace_id = OLD.workspace_id AND role = 'admin' AND id <> OLD.id
        ) THEN
            RAISE EXCEPTION 'nexo:last_admin' USING ERRCODE = 'check_violation';
        END IF;
    END IF;
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;
    RETURN NEW;
END
$$;
-- +goose StatementEnd

CREATE TRIGGER memberships_keep_one_admin
    BEFORE UPDATE OF role OR DELETE ON memberships
    FOR EACH ROW EXECUTE FUNCTION nexo_keep_one_admin();

ALTER TABLE workspaces ENABLE ROW LEVEL SECURITY;
CREATE POLICY workspaces_select_member ON workspaces FOR SELECT
    USING (nexo_is_member(id));
CREATE POLICY workspaces_update_admin ON workspaces FOR UPDATE
    USING (id = nexo_workspace_id() AND nexo_workspace_role(id) = 'admin')
    WITH CHECK (id = nexo_workspace_id());

ALTER TABLE memberships ENABLE ROW LEVEL SECURITY;
CREATE POLICY memberships_select ON memberships FOR SELECT
    USING (user_id = nexo_user_id()
           OR (workspace_id = nexo_workspace_id() AND nexo_is_member(workspace_id)));
CREATE POLICY memberships_update_admin ON memberships FOR UPDATE
    USING (workspace_id = nexo_workspace_id() AND nexo_workspace_role(workspace_id) = 'admin')
    WITH CHECK (workspace_id = nexo_workspace_id());
CREATE POLICY memberships_delete ON memberships FOR DELETE
    USING (workspace_id = nexo_workspace_id()
           AND (nexo_workspace_role(workspace_id) = 'admin' OR user_id = nexo_user_id()));

CREATE POLICY users_select_coworkers ON users FOR SELECT
    USING (nexo_shares_workspace(id));

-- Column privileges: RLS limits rows, these limit what can change.
REVOKE INSERT, UPDATE, DELETE ON workspaces FROM nexo_app;
GRANT UPDATE (name, updated_at, deleted_at) ON workspaces TO nexo_app;
REVOKE INSERT, UPDATE ON memberships FROM nexo_app;
GRANT UPDATE (role, updated_at) ON memberships TO nexo_app;

REVOKE ALL ON FUNCTION nexo_workspace_role(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION nexo_is_member(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION nexo_shares_workspace(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION nexo_create_workspace(uuid, uuid, text) FROM PUBLIC;
REVOKE ALL ON FUNCTION nexo_keep_one_admin() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION nexo_workspace_role(uuid) TO nexo_app;
GRANT EXECUTE ON FUNCTION nexo_is_member(uuid) TO nexo_app;
GRANT EXECUTE ON FUNCTION nexo_shares_workspace(uuid) TO nexo_app;
GRANT EXECUTE ON FUNCTION nexo_create_workspace(uuid, uuid, text) TO nexo_app;

-- +goose Down
DROP POLICY IF EXISTS users_select_coworkers ON users;
DROP TABLE IF EXISTS memberships;
DROP TABLE IF EXISTS workspaces;
DROP FUNCTION IF EXISTS nexo_keep_one_admin();
DROP FUNCTION IF EXISTS nexo_create_workspace(uuid, uuid, text);
DROP FUNCTION IF EXISTS nexo_shares_workspace(uuid);
DROP FUNCTION IF EXISTS nexo_is_member(uuid);
DROP FUNCTION IF EXISTS nexo_workspace_role(uuid);
