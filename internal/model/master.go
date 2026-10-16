package model

import (
	"time"

	"github.com/google/uuid"
)

type ProductCategory struct {
	Base
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	IsActive    bool    `json:"is_active"`
}

type Product struct {
	Base
	ProductCategoryID uuid.UUID `json:"product_category_id"`
	SKU               string    `gorm:"column:sku" json:"sku"`
	Name              string    `json:"name"`
	Description       *string   `json:"description"`
	Unit              string    `json:"unit"`
	MinimumStock      int       `json:"minimum_stock"`
	IsActive          bool      `json:"is_active"`

	ProductCategory *ProductCategory `json:"product_category"`
}

type Warehouse struct {
	Base
	Code     string  `json:"code"`
	Name     string  `json:"name"`
	Address  *string `json:"address"`
	IsActive bool    `json:"is_active"`
}

type WarehouseLocation struct {
	Base
	WarehouseID uuid.UUID `json:"warehouse_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	IsActive    bool      `json:"is_active"`

	Warehouse *Warehouse `json:"warehouse"`
}

type AuditTrail struct {
	ID          uuid.UUID `gorm:"type:uuid;default:get_random_uuid();primaryKey" json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Action      string    `json:"action"`
	EntityType  string    `json:"entity_type"`
	EntityID    uuid.UUID `json:"entity_id"`
	Description *string   `json:"description"`
	OccurredAt  time.Time `json:"occurred_at"`
	CreatedAt   time.Time `json:"created_at"`

	User User `json:"user"`
}
