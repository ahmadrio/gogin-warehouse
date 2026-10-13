CREATE TABLE operations (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name        varchar(50) NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);
CREATE UNIQUE INDEX uq_operations_name ON operations (name) WHERE deleted_at IS NULL;

CREATE TABLE resources (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name        varchar(50) NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);
CREATE UNIQUE INDEX uq_resources_name ON resources (name) WHERE deleted_at IS NULL;

CREATE TABLE permissions (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    operation_id  uuid NOT NULL REFERENCES operations (id),
    resource_id   uuid NOT NULL REFERENCES resources (id),
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    deleted_at    timestamptz
);
CREATE INDEX idx_permissions_operation_id ON permissions (operation_id);
CREATE INDEX idx_permissions_resource_id ON permissions (resource_id);
CREATE UNIQUE INDEX uq_permissions_operation_resource
    ON permissions (operation_id, resource_id) WHERE deleted_at IS NULL;

CREATE TABLE roles (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name         varchar(50) NOT NULL,
    description  text,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    deleted_at   timestamptz
);
CREATE UNIQUE INDEX uq_roles_name ON roles (name) WHERE deleted_at IS NULL;

CREATE TABLE role_permissions (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id        uuid NOT NULL REFERENCES roles (id),
    permission_id  uuid NOT NULL REFERENCES permissions (id),
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    deleted_at     timestamptz
);
CREATE INDEX idx_role_permissions_role_id ON role_permissions (role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions (permission_id);
CREATE UNIQUE INDEX uq_role_permissions
    ON role_permissions (role_id, permission_id) WHERE deleted_at IS NULL;

CREATE TABLE users (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email       varchar(255) NOT NULL,
    name        varchar(255) NOT NULL,
    password    varchar(255) NOT NULL,
    is_active   boolean NOT NULL DEFAULT true,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);
CREATE UNIQUE INDEX uq_users_email ON users (lower(email)) WHERE deleted_at IS NULL;

CREATE TABLE user_roles (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid NOT NULL REFERENCES users (id),
    role_id     uuid NOT NULL REFERENCES roles (id),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);
CREATE INDEX idx_user_roles_user_id ON user_roles (user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles (role_id);
CREATE UNIQUE INDEX uq_user_roles ON user_roles (user_id, role_id) WHERE deleted_at IS NULL;
