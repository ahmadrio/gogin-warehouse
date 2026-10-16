package model

import (
	"time"

	"github.com/google/uuid"
)

// Nilai kolom stock_movements.movement_type (juga dipakai sebagai source_entity_type)
const (
	MovementInbound         = "inbound"
	MovementOutbound        = "outbound"
	MovementStockTransfer   = "stock_transfer"
	MovementStockAdjustment = "stock_adjustment"
)

// Nilai kolom stock_movements.movement_direction
const (
	DirectionIn  = "in"
	DirectionOut = "out"
)

// WarehouseLocationItem = SALDO stok satu produk di satu lokasi (rak).
// Tidak ada DeletedAt: saldo tidak dihapus, cukup bernilai 0.
type WarehouseLocationItem struct {
	ID                  uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WarehouseID         uuid.UUID `json:"warehouse_id"`
	WarehouseLocationID uuid.UUID `json:"warehouse_location_id"`
	ProductID           uuid.UUID `json:"product_id"`
	Quantity            int       `json:"quantity"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// StockMovement = satu baris LEDGER. Hanya pernah di-INSERT, tidak pernah diubah atau dihapus.
type StockMovement struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WarehouseID         uuid.UUID  `json:"warehouse_id"`
	WarehouseLocationID uuid.UUID  `json:"warehouse_location_id"`
	ProductID           uuid.UUID  `json:"product_id"`
	MovementType        string     `json:"movement_type"`
	MovementDirection   string     `json:"movement_direction"`
	Quantity            int        `json:"quantity"`
	StockAfter          int        `json:"stock_after"`
	SourceEntityType    string     `json:"source_entity_type"`
	SourceEntityID      uuid.UUID  `json:"source_entity_id"`
	SourceItemID        *uuid.UUID `json:"source_item_id"`
	OccurredAt          time.Time  `json:"occurred_at"`
	CreatedByUserID     uuid.UUID  `json:"created_by_user_id"`
	CreatedAt           time.Time  `json:"created_at"`
}
