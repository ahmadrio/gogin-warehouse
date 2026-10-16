package model

import (
	"time"

	"github.com/google/uuid"
)

// Status transaksi yang butuh approval (dipakai juga oleh In-Bound, Out-Bound, Transfer nanti)
const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
)

// Tipe stock adjustment
const (
	AdjustmentStockOpname  = "stock_opname"  // hasil hitung fisik, boleh + atau -
	AdjustmentDamagedGoods = "damaged_goods" // barang rusak/cacat, hanya -
	AdjustmentLostGoods    = "lost_goods"    // barang hilang, hanya -
	AdjustmentCorrection   = "correction"    // koreksi salah catat, boleh + atau -
)

type StockAdjustment struct {
	Base
	WarehouseID      uuid.UUID  `json:"warehouse_id"`
	ReferenceNumber  string     `json:"reference_number"`
	Status           string     `json:"status"`
	AdjustmentType   string     `json:"adjustment_type"`
	Reason           string     `json:"reason"`
	RejectionNote    *string    `json:"rejection_note"`
	SubmittedAt      *time.Time `json:"submitted_at"`
	ReviewedAt       *time.Time `json:"reviewed_at"`
	CreatedByUserID  uuid.UUID  `json:"created_by_user_id"`
	ReviewedByUserID *uuid.UUID `json:"reviewed_by_user_id"`

	// has-many: GORM mencocokkan lewat kolom stock_adjustment_id di tabel item
	Items []StockAdjustmentItem `json:"items,omitempty"`
}

// StockAdjustmentItem tidak memakai Base karena tabelnya tidak punya deleted_at.
type StockAdjustmentItem struct {
	ID                  uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	StockAdjustmentID   uuid.UUID `json:"stock_adjustment_id"`
	ProductID           uuid.UUID `json:"product_id"`
	WarehouseLocationID uuid.UUID `json:"warehouse_location_id"`
	QuantityChange      int       `json:"quantity_change"`
	Notes               *string   `json:"notes"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`

	// belongs-to, diisi dengan Preload("Items.Product") dan Preload("Items.WarehouseLocation")
	Product           *Product           `json:"product,omitempty"`
	WarehouseLocation *WarehouseLocation `json:"warehouse_location,omitempty"`
}
