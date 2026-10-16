package dto

import "github.com/google/uuid"

type InboundRequest struct {
	WarehouseID         uuid.UUID            `json:"warehouse_id"` // hanya dibaca saat create
	PurchaseOrderNumber *string              `json:"purchase_order_number" binding:"omitempty,max=100"`
	Notes               *string              `json:"notes"`
	Items               []InboundItemRequest `json:"items" binding:"required,min=1,max=100,dive"`
}

type InboundItemRequest struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"gt=0"`
	Notes     *string   `json:"notes"`
}

// PutawayRequest: rak tujuan untuk SETIAP item In-Bound.
type PutawayRequest struct {
	Items []PutawayItemRequest `json:"items" binding:"required,min=1,dive"`
}

type PutawayItemRequest struct {
	InboundItemID       uuid.UUID `json:"inbound_item_id" binding:"required"`
	WarehouseLocationID uuid.UUID `json:"warehouse_location_id" binding:"required"`
}
