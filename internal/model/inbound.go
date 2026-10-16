package model

import (
	"time"

	"github.com/google/uuid"
)

// StatusCompleted: In-Bound sudah di-putaway ke rak dan stok sudah bertambah.
const StatusCompleted = "completed"

type Inbound struct {
	Base
	WarehouseID         uuid.UUID  `json:"warehouse_id"`
	ReferenceNumber     string     `json:"reference_number"`
	PurchaseOrderNumber *string    `json:"purchase_order_number"`
	Status              string     `json:"status"`
	Notes               *string    `json:"notes"`
	RejectionNote       *string    `json:"rejection_note"`
	SubmittedAt         *time.Time `json:"submitted_at"`
	ReviewedAt          *time.Time `json:"reviewed_at"`
	PutawayAt           *time.Time `json:"putaway_at"`
	CreatedByUserID     uuid.UUID  `json:"created_by_user_id"`
	ReviewedByUserID    *uuid.UUID `json:"reviewed_by_user_id"`
	PutawayByUserID     *uuid.UUID `json:"putaway_by_user_id"`

	Items []InboundItem `json:"items,omitempty"`
}

type InboundItem struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	InboundID           uuid.UUID  `json:"inbound_id"`
	ProductID           uuid.UUID  `json:"product_id"`
	Quantity            int        `json:"quantity"`
	WarehouseLocationID *uuid.UUID `json:"warehouse_location_id"` // nil sampai putaway
	Notes               *string    `json:"notes"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`

	Product           *Product           `json:"product,omitempty"`
	WarehouseLocation *WarehouseLocation `json:"warehouse_location,omitempty"`
}
