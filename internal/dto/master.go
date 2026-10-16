package dto

import "github.com/google/uuid"

type CreateProductRequest struct {
	ProductCategoryID uuid.UUID `json:"product_category_id" binding:"required"`
	SKU               string    `json:"sku" binding:"required,max=100"`
	Name              string    `json:"name" binding:"required,max=255"`
	Description       *string   `json:"description"`
	Unit              string    `json:"unit" binding:"required,oneof=pcs box kg liter"`
	MinimumStock      int       `json:"minimum_stock" binding:"gte=0"`
	IsActive          *bool     `json:"is_active"` // default true jika tidak dikirim
}

type UpdateProductRequest struct {
	ProductCategoryID uuid.UUID `json:"product_category_id" binding:"required"`
	SKU               string    `json:"sku" binding:"required,max=100"`
	Name              string    `json:"name" binding:"required,max=255"`
	Description       *string   `json:"description"`
	Unit              string    `json:"unit" binding:"required,oneof=pcs box kg liter"`
	MinimumStock      int       `json:"minimum_stock" binding:"gte=0"`
	IsActive          *bool     `json:"is_active" binding:"required"`
}

type CreateProductCategoryRequest struct {
	Code        string  `json:"code" binding:"required,max=50"`
	Name        string  `json:"name" binding:"required,max=255"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"` // default true jika tidak dikirim
}

type UpdateProductCategoryRequest struct {
	Code        string  `json:"code" binding:"required,max=50"`
	Name        string  `json:"name" binding:"required,max=255"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active" binding:"required"`
}

type CreateWarehouseRequest struct {
	Code    string  `json:"code" binding:"required,max=50"`
	Name    string  `json:"name" binding:"required,max=255"`
	Address *string `json:"address"`
}

type CreateWarehouseLocationRequest struct {
	Code        string  `json:"code" binding:"required,max=50"`
	Name        string  `json:"name" binding:"required,max=255"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"` // default true jika tidak dikirim
}

type UpdateWarehouseLocationRequest struct {
	Code        string  `json:"code" binding:"required,max=50"`
	Name        string  `json:"name" binding:"required,max=255"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active" binding:"required"`
}
