package model

import (
	"time"

	"github.com/google/uuid"
)

type StockTransfer struct {
	Base
	WarehouseID      uuid.UUID  `json:"warehouse_id"`
	ReferenceNumber  string     `json:"reference_number"`
	Status           string     `json:"status"`
	Notes            *string    `json:"notes"`
	RejectionNote    *string    `json:"rejection_note"`
	SubmittedAt      *time.Time `json:"submitted_at"`
	ReviewedAt       *time.Time `json:"reviewed_at"`
	CreatedByUserID  uuid.UUID  `json:"created_by_user_id"`
	ReviewedByUserID *uuid.UUID `json:"reviewed_by_user_id"`

	Items []StockTransferItem `json:"items,omitempty"`
}

type StockTransferItem struct {
	ID                             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	StockTransferID                uuid.UUID `json:"stock_transfer_id"`
	ProductID                      uuid.UUID `json:"product_id"`
	SourceWarehouseLocationID      uuid.UUID `json:"source_warehouse_location_id"`
	DestinationWarehouseLocationID uuid.UUID `json:"destination_warehouse_location_id"`
	Quantity                       int       `json:"quantity"`
	Notes                          *string   `json:"notes"`
	CreatedAt                      time.Time `json:"created_at"`
	UpdatedAt                      time.Time `json:"updated_at"`

	Product *Product `json:"product,omitempty"`
	// Dua relasi ke tabel yang sama -> foreign key-nya harus disebut eksplisit.
	SourceLocation      *WarehouseLocation `gorm:"foreignKey:SourceWarehouseLocationID" json:"source_location,omitempty"`
	DestinationLocation *WarehouseLocation `gorm:"foreignKey:DestinationWarehouseLocationID" json:"destination_location,omitempty"`
}
