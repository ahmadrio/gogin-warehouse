package handler

import (
	"github.com/gin-gonic/gin"

	"warehouse-management/internal/auth"
	"warehouse-management/internal/dto"
	"warehouse-management/internal/repository"
	"warehouse-management/internal/service"
	"warehouse-management/pkg/pagination"
	"warehouse-management/pkg/response"
)

type ProductHandler struct {
	service *service.ProductService
}

func NewProductHandler(s *service.ProductService) *ProductHandler {
	return &ProductHandler{service: s}
}

// GET /api/v1/products?category_id=&search=&page=
func (h *ProductHandler) List(c *gin.Context) {
	categoryID, err := queryUUID(c, "category_id")
	if err != nil {
		response.Error(c, err)
		return
	}
	p := pagination.FromQuery(c)
	items, total, err := h.service.List(c.Request.Context(), repository.ProductFilter{CategoryID: categoryID}, p)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, items, pagination.NewMeta(p, total))
}

// GET /api/v1/products/:id
func (h *ProductHandler) Get(c *gin.Context) {
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

// POST /api/v1/products
func (h *ProductHandler) Create(c *gin.Context) {
	var req dto.CreateProductRequest
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

// PUT /api/v1/products/:id
func (h *ProductHandler) Update(c *gin.Context) {
	id, err := paramUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.UpdateProductRequest
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

// DELETE /api/v1/products/:id
func (h *ProductHandler) Delete(c *gin.Context) {
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
