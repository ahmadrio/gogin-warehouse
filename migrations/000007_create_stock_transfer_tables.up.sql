CREATE TABLE stock_transfers (
    id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    warehouse_id         uuid NOT NULL REFERENCES warehouses (id),
    reference_number     varchar(30) NOT NULL UNIQUE,   -- TRF-202610-0001
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
CREATE INDEX idx_stock_transfers_warehouse_status ON stock_transfers (warehouse_id, status);
CREATE INDEX idx_stock_transfers_created_by ON stock_transfers (created_by_user_id);
CREATE INDEX idx_stock_transfers_created_at ON stock_transfers (created_at);

CREATE TABLE stock_transfer_items (
    id                                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    stock_transfer_id                  uuid NOT NULL REFERENCES stock_transfers (id),
    product_id                         uuid NOT NULL REFERENCES products (id),
    source_warehouse_location_id       uuid NOT NULL REFERENCES warehouse_locations (id),
    destination_warehouse_location_id  uuid NOT NULL REFERENCES warehouse_locations (id),
    quantity                           int NOT NULL CHECK (quantity > 0),
    notes                              text,
    created_at                         timestamptz NOT NULL DEFAULT now(),
    updated_at                         timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT chk_transfer_different_location
        CHECK (source_warehouse_location_id <> destination_warehouse_location_id),
    CONSTRAINT uq_stock_transfer_items
        UNIQUE (stock_transfer_id, product_id, source_warehouse_location_id, destination_warehouse_location_id)
);
CREATE INDEX idx_sti_product ON stock_transfer_items (product_id);
CREATE INDEX idx_sti_source ON stock_transfer_items (source_warehouse_location_id);
CREATE INDEX idx_sti_destination ON stock_transfer_items (destination_warehouse_location_id);
