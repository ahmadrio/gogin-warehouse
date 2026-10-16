-- document_sequences       : penomoran dokumen (ADJ-202610-0001), BR-003
-- warehouse_location_items : SALDO stok per (lokasi, produk)
-- stock_movements          : LEDGER, satu baris per perubahan stok, BR-006
-- Kedua tabel stok tidak memakai soft delete: saldo cukup bernilai 0,
-- dan ledger tidak boleh diubah atau dihapus (BR-009).

CREATE TABLE document_sequences (
    prefix       varchar(20) NOT NULL,   -- ADJ, INB, OUT, TRF
    period       varchar(6)  NOT NULL,   -- YYYYMM
    last_number  int NOT NULL,
    updated_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (prefix, period)
);

CREATE TABLE warehouse_location_items (
    id                     uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    warehouse_id           uuid NOT NULL REFERENCES warehouses (id),
    warehouse_location_id  uuid NOT NULL REFERENCES warehouse_locations (id),
    product_id             uuid NOT NULL REFERENCES products (id),
    quantity               int NOT NULL DEFAULT 0 CHECK (quantity >= 0),   -- BR-005
    created_at             timestamptz NOT NULL DEFAULT now(),
    updated_at             timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uq_warehouse_location_items UNIQUE (warehouse_location_id, product_id)
);
CREATE INDEX idx_wli_warehouse_id ON warehouse_location_items (warehouse_id);
CREATE INDEX idx_wli_product_id ON warehouse_location_items (product_id);
CREATE INDEX idx_wli_warehouse_product ON warehouse_location_items (warehouse_id, product_id);

CREATE TABLE stock_movements (
    id                     uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    warehouse_id           uuid NOT NULL REFERENCES warehouses (id),
    warehouse_location_id  uuid NOT NULL REFERENCES warehouse_locations (id),
    product_id             uuid NOT NULL REFERENCES products (id),
    movement_type          varchar(30) NOT NULL,   -- inbound, outbound, stock_transfer, stock_adjustment
    movement_direction     varchar(3)  NOT NULL CHECK (movement_direction IN ('in', 'out')),
    quantity               int NOT NULL CHECK (quantity > 0),
    stock_after            int NOT NULL CHECK (stock_after >= 0),
    source_entity_type     varchar(30) NOT NULL,
    source_entity_id       uuid NOT NULL,
    source_item_id         uuid,
    occurred_at            timestamptz NOT NULL DEFAULT now(),
    created_by_user_id     uuid NOT NULL REFERENCES users (id),
    created_at             timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_sm_warehouse_product_time ON stock_movements (warehouse_id, product_id, occurred_at);
CREATE INDEX idx_sm_location_product_time ON stock_movements (warehouse_location_id, product_id, occurred_at);
CREATE INDEX idx_sm_source ON stock_movements (source_entity_type, source_entity_id);
CREATE INDEX idx_sm_occurred_at ON stock_movements (occurred_at);
