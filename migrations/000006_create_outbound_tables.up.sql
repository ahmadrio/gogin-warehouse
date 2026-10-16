CREATE TABLE outbounds (
    id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    warehouse_id         uuid NOT NULL REFERENCES warehouses (id),
    reference_number     varchar(30) NOT NULL UNIQUE,   -- OUT-202610-0001
    status               varchar(20) NOT NULL CHECK (status IN ('pending', 'approved', 'rejected')),
    notes                text,
    rejection_note       text,
    submitted_at         timestamptz,
    reviewed_at          timestamptz,
    created_by_user_id   uuid NOT NULL REFERENCES users (id),
    reviewed_by_user_id  uuid REFERENCES users (id),
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    deleted_at           timestamptz
);
CREATE INDEX idx_outbounds_warehouse_status ON outbounds (warehouse_id, status);
CREATE INDEX idx_outbounds_created_by ON outbounds (created_by_user_id);
CREATE INDEX idx_outbounds_created_at ON outbounds (created_at);

CREATE TABLE outbound_items (
    id                     uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    outbound_id            uuid NOT NULL REFERENCES outbounds (id),
    product_id             uuid NOT NULL REFERENCES products (id),
    warehouse_location_id  uuid NOT NULL REFERENCES warehouse_locations (id),
    quantity               int NOT NULL CHECK (quantity > 0),
    notes                  text,
    created_at             timestamptz NOT NULL DEFAULT now(),
    updated_at             timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uq_outbound_items UNIQUE (outbound_id, warehouse_location_id, product_id)
);
CREATE INDEX idx_outbound_items_product ON outbound_items (product_id);
CREATE INDEX idx_outbound_items_location ON outbound_items (warehouse_location_id);
