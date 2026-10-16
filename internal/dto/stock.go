package dto

import "github.com/google/uuid"

// StockAdjustmentRequest dipakai untuk create (POST) dan revisi (PUT).
// Warehouse tidak bisa diganti saat revisi, jadi warehouse_id hanya dibaca saat create.
type StockAdjustmentRequest struct {
	WarehouseID    uuid.UUID                    `json:"warehouse_id"`
	AdjustmentType string                       `json:"adjustment_type" binding:"required,oneof=stock_opname damaged_goods lost_goods correction"`
	Reason         string                       `json:"reason" binding:"required,max=1000"`
	Items          []StockAdjustmentItemRequest `json:"items" binding:"required,min=1,max=100,dive"`
}

type StockAdjustmentItemRequest struct {
	ProductID           uuid.UUID `json:"product_id" binding:"required"`
	WarehouseLocationID uuid.UUID `json:"warehouse_location_id" binding:"required"`
	QuantityChange      int       `json:"quantity_change" binding:"ne=0"`
	Notes               *string   `json:"notes"`
}

type RejectRequest struct {
	RejectionNote string `json:"rejection_note" binding:"required,max=1000"`
}
