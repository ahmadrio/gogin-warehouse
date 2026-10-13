CREATE TABLE product_categories (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code         varchar(50) NOT NULL,
    name         varchar(255) NOT NULL,
    description  text,
    is_active    boolean NOT NULL DEFAULT true,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    deleted_at   timestamptz
);
CREATE UNIQUE INDEX uq_product_categories_code ON product_categories (code) WHERE deleted_at IS NULL;

CREATE TABLE products (
    id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    product_category_id  uuid NOT NULL REFERENCES product_categories (id),
    sku                  varchar(100) NOT NULL,
    name                 varchar(255) NOT NULL,
    description          text,
    unit                 varchar(20) NOT NULL,
    minimum_stock        int NOT NULL DEFAULT 0 CHECK (minimum_stock >= 0),
    is_active            boolean NOT NULL DEFAULT true,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    deleted_at           timestamptz
);
CREATE INDEX idx_products_category_id ON products (product_category_id);
CREATE INDEX idx_products_name ON products (name);
CREATE UNIQUE INDEX uq_products_sku ON products (sku) WHERE deleted_at IS NULL;

CREATE TABLE warehouses (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code        varchar(50) NOT NULL,
    name        varchar(255) NOT NULL,
    address     text,
    is_active   boolean NOT NULL DEFAULT true,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);
CREATE INDEX idx_warehouses_name ON warehouses (name);
CREATE UNIQUE INDEX uq_warehouses_code ON warehouses (code) WHERE deleted_at IS NULL;

CREATE TABLE warehouse_locations (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    warehouse_id  uuid NOT NULL REFERENCES warehouses (id),
    code          varchar(50) NOT NULL,
    name          varchar(255) NOT NULL,
    description   text,
    is_active     boolean NOT NULL DEFAULT true,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    deleted_at    timestamptz
);
CREATE INDEX idx_warehouse_locations_warehouse_id ON warehouse_locations (warehouse_id);
CREATE UNIQUE INDEX uq_warehouse_locations_code
    ON warehouse_locations (warehouse_id, code) WHERE deleted_at IS NULL;

CREATE TABLE user_warehouses (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       uuid NOT NULL REFERENCES users (id),
    warehouse_id  uuid NOT NULL REFERENCES warehouses (id),
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    deleted_at    timestamptz
);
CREATE INDEX idx_user_warehouses_user_id ON user_warehouses (user_id);
CREATE INDEX idx_user_warehouses_warehouse_id ON user_warehouses (warehouse_id);
CREATE UNIQUE INDEX uq_user_warehouses
    ON user_warehouses (user_id, warehouse_id) WHERE deleted_at IS NULL;
