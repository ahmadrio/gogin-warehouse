package dto

import "github.com/google/uuid"

type OutboundRequest struct {
	WarehouseID uuid.UUID             `json:"warehouse_id"` // hanya dibaca saat create
	Notes       *string               `json:"notes"`
	Items       []OutboundItemRequest `json:"items" binding:"required,min=1,max=100,dive"`
}

type OutboundItemRequest struct {
	ProductID           uuid.UUID `json:"product_id" binding:"required"`
	WarehouseLocationID uuid.UUID `json:"warehouse_location_id" binding:"required"`
	Quantity            int       `json:"quantity" binding:"gt=0"`
	Notes               *string   `json:"notes"`
}
