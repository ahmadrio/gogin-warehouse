package model

import (
	"time"

	"github.com/google/uuid"
)

type Outbound struct {
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

	Items []OutboundItem `json:"items,omitempty"`
}

type OutboundItem struct {
	ID                  uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OutboundID          uuid.UUID `json:"outbound_id"`
	ProductID           uuid.UUID `json:"product_id"`
	WarehouseLocationID uuid.UUID `json:"warehouse_location_id"`
	Quantity            int       `json:"quantity"`
	Notes               *string   `json:"notes"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`

	Product           *Product           `json:"product,omitempty"`
	WarehouseLocation *WarehouseLocation `json:"warehouse_location,omitempty"`
}
