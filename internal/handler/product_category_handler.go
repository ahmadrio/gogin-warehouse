package handler

import (
	"github.com/gin-gonic/gin"

	"warehouse-management/internal/auth"
	"warehouse-management/internal/dto"
	"warehouse-management/internal/service"
	"warehouse-management/pkg/pagination"
	"warehouse-management/pkg/response"
)

type ProductCategoryHandler struct {
	service *service.ProductCategoryService
}

func NewProductCategoryHandler(s *service.ProductCategoryService) *ProductCategoryHandler {
	return &ProductCategoryHandler{service: s}
}

// GET /api/v1/product-categories?page=1&per_page=15&search=elek
func (h *ProductCategoryHandler) List(c *gin.Context) {
	p := pagination.FromQuery(c)
	items, total, err := h.service.List(c.Request.Context(), p)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, items, pagination.NewMeta(p, total))
}

// GET /api/v1/product-categories/:id
func (h *ProductCategoryHandler) Get(c *gin.Context) {
	id, err := paramUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	item, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, item)
}

// POST /api/v1/product-categories
func (h *ProductCategoryHandler) Create(c *gin.Context) {
	var req dto.CreateProductCategoryRequest
	if err := bindJSON(c, &req); err != nil {
		response.Error(c, err)
		return
	}
	item, err := h.service.Create(c.Request.Context(), auth.CurrentUser(c), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, item)
}

// PUT /api/v1/product-categories/:id
func (h *ProductCategoryHandler) Update(c *gin.Context) {
	id, err := paramUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.UpdateProductCategoryRequest
	if err := bindJSON(c, &req); err != nil {
		response.Error(c, err)
		return
	}
	item, err := h.service.Update(c.Request.Context(), auth.CurrentUser(c), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, item)
}

// DELETE /api/v1/product-categories/:id
func (h *ProductCategoryHandler) Delete(c *gin.Context) {
	id, err := paramUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.service.Delete(c.Request.Context(), auth.CurrentUser(c), id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}
