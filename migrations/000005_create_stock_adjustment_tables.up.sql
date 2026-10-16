CREATE TABLE stock_adjustments (
    id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    warehouse_id         uuid NOT NULL REFERENCES warehouses (id),
    reference_number     varchar(30) NOT NULL UNIQUE,   -- ADJ-202610-0001, tidak pernah dipakai ulang
    status               varchar(20) NOT NULL CHECK (status IN ('pending', 'approved', 'rejected')),
    adjustment_type      varchar(30) NOT NULL CHECK (adjustment_type IN ('stock_opname', 'damaged_goods', 'lost_goods', 'correction')),
    reason               text NOT NULL,
    rejection_note       text,
    submitted_at         timestamptz,
    reviewed_at          timestamptz,
    created_by_user_id   uuid NOT NULL REFERENCES users (id),
    reviewed_by_user_id  uuid REFERENCES users (id),
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    deleted_at           timestamptz
);
CREATE INDEX idx_sa_warehouse_status ON stock_adjustments (warehouse_id, status);
CREATE INDEX idx_sa_warehouse_type ON stock_adjustments (warehouse_id, adjustment_type);
CREATE INDEX idx_sa_created_by ON stock_adjustments (created_by_user_id);
CREATE INDEX idx_sa_created_at ON stock_adjustments (created_at);

-- Item tidak memakai soft delete: saat direvisi, item lama dihapus dan diganti.
CREATE TABLE stock_adjustment_items (
    id                     uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    stock_adjustment_id    uuid NOT NULL REFERENCES stock_adjustments (id),
    product_id             uuid NOT NULL REFERENCES products (id),
    warehouse_location_id  uuid NOT NULL REFERENCES warehouse_locations (id),
    quantity_change        int NOT NULL CHECK (quantity_change <> 0),   -- positif menambah, negatif mengurangi
    notes                  text,
    created_at             timestamptz NOT NULL DEFAULT now(),
    updated_at             timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uq_sa_items UNIQUE (stock_adjustment_id, warehouse_location_id, product_id)
);
CREATE INDEX idx_sa_items_product ON stock_adjustment_items (product_id);
CREATE INDEX idx_sa_items_location ON stock_adjustment_items (warehouse_location_id);
