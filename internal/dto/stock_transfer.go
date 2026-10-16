package dto

import "github.com/google/uuid"

type StockTransferRequest struct {
	WarehouseID uuid.UUID                  `json:"warehouse_id"` // hanya dibaca saat create
	Notes       *string                    `json:"notes"`
	Items       []StockTransferItemRequest `json:"items" binding:"required,min=1,max=100,dive"`
}

type StockTransferItemRequest struct {
	ProductID                      uuid.UUID `json:"product_id" binding:"required"`
	SourceWarehouseLocationID      uuid.UUID `json:"source_warehouse_location_id" binding:"required"`
	DestinationWarehouseLocationID uuid.UUID `json:"destination_warehouse_location_id" binding:"required"`
	Quantity                       int       `json:"quantity" binding:"gt=0"`
	Notes                          *string   `json:"notes"`
}
