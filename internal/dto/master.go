package dto

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
