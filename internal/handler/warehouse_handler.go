package handler

import (
	"github.com/gin-gonic/gin"

	"warehouse-management/internal/auth"
	"warehouse-management/internal/dto"
	"warehouse-management/internal/service"
	"warehouse-management/pkg/pagination"
	"warehouse-management/pkg/response"
)

type WarehouseHandler struct {
	service *service.WarehouseService
}

func NewWarehouseHandler(s *service.WarehouseService) *WarehouseHandler {
	return &WarehouseHandler{service: s}
}

// GET /api/v1/warehouses
func (h *WarehouseHandler) List(c *gin.Context) {
	p := pagination.FromQuery(c)
	items, total, err := h.service.List(c.Request.Context(), auth.CurrentUser(c), p)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, items, pagination.NewMeta(p, total))
}

// GET /api/v1/warehouses/:id
func (h *WarehouseHandler) Get(c *gin.Context) {
	id, err := paramUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	item, err := h.service.Get(c.Request.Context(), auth.CurrentUser(c), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, item)
}

// GET /api/v1/warehouses/:id/locations
func (h *WarehouseHandler) ListLocations(c *gin.Context) {
	id, err := paramUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	items, err := h.service.ListLocations(c.Request.Context(), auth.CurrentUser(c), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, items)
}

// POST /api/v1/warehouses
func (h *WarehouseHandler) Create(c *gin.Context) {
	var req dto.CreateWarehouseRequest
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
