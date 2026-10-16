-- Perbedaan dari DBML:
--   * status punya nilai tambahan 'completed' (sudah di-putaway)
--   * inbound_items.warehouse_location_id (rak tujuan, diisi saat putaway)
--   * inbounds.putaway_at & putaway_by_user_id (siapa & kapan putaway, BR-010)
CREATE TABLE inbounds (
    id                     uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    warehouse_id           uuid NOT NULL REFERENCES warehouses (id),
    reference_number       varchar(30) NOT NULL UNIQUE,   -- INB-202610-0001
    purchase_order_number  varchar(100),
    status                 varchar(20) NOT NULL CHECK (status IN ('pending', 'approved', 'rejected', 'completed')),
    notes                  text,
    rejection_note         text,
    submitted_at           timestamptz,
    reviewed_at            timestamptz,
    putaway_at             timestamptz,
    created_by_user_id     uuid NOT NULL REFERENCES users (id),
    reviewed_by_user_id    uuid REFERENCES users (id),
    putaway_by_user_id     uuid REFERENCES users (id),
    created_at             timestamptz NOT NULL DEFAULT now(),
    updated_at             timestamptz NOT NULL DEFAULT now(),
    deleted_at             timestamptz
);
CREATE INDEX idx_inbounds_warehouse_status ON inbounds (warehouse_id, status);
CREATE INDEX idx_inbounds_created_by ON inbounds (created_by_user_id);
CREATE INDEX idx_inbounds_created_at ON inbounds (created_at);

CREATE TABLE inbound_items (
    id                     uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    inbound_id             uuid NOT NULL REFERENCES inbounds (id),
    product_id             uuid NOT NULL REFERENCES products (id),
    quantity               int NOT NULL CHECK (quantity > 0),
    warehouse_location_id  uuid REFERENCES warehouse_locations (id),   -- NULL sampai putaway
    notes                  text,
    created_at             timestamptz NOT NULL DEFAULT now(),
    updated_at             timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uq_inbound_items UNIQUE (inbound_id, product_id)
);
CREATE INDEX idx_inbound_items_product ON inbound_items (product_id);
