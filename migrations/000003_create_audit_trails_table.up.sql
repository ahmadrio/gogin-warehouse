CREATE TABLE audit_trails (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      uuid NOT NULL REFERENCES users (id),
    action       varchar(50) NOT NULL,
    entity_type  varchar(50) NOT NULL,
    entity_id    uuid NOT NULL,
    description  text,
    occurred_at  timestamptz NOT NULL DEFAULT now(),
    created_at   timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_audit_trails_user_id ON audit_trails (user_id);
CREATE INDEX idx_audit_trails_entity ON audit_trails (entity_type, entity_id);
CREATE INDEX idx_audit_trails_action ON audit_trails (action);
CREATE INDEX idx_audit_trails_occurred_at ON audit_trails (occurred_at);
