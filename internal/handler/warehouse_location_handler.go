package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"warehouse-management/internal/auth"
	"warehouse-management/internal/dto"
	"warehouse-management/internal/service"
	"warehouse-management/pkg/response"
)

// List lokasi sudah ada di WarehouseHandler.ListLocations (Tahap 11).
// Handler ini menambah detail, create, update, dan delete.
type WarehouseLocationHandler struct {
	service *service.WarehouseLocationService
}

func NewWarehouseLocationHandler(s *service.WarehouseLocationService) *WarehouseLocationHandler {
	return &WarehouseLocationHandler{service: s}
}

// locationIDs membaca :id (warehouse) dan :location_id dari URL sekaligus.
func locationIDs(c *gin.Context) (uuid.UUID, uuid.UUID, error) {
	warehouseID, err := paramUUID(c, "id")
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	locationID, err := paramUUID(c, "location_id")
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return warehouseID, locationID, nil
}

// GET /api/v1/warehouses/:id/locations/:location_id
func (h *WarehouseLocationHandler) Get(c *gin.Context) {
	warehouseID, locationID, err := locationIDs(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	item, err := h.service.Get(c.Request.Context(), auth.CurrentUser(c), warehouseID, locationID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, item)
}

// POST /api/v1/warehouses/:id/locations
func (h *WarehouseLocationHandler) Create(c *gin.Context) {
	warehouseID, err := paramUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.CreateWarehouseLocationRequest
	if err := bindJSON(c, &req); err != nil {
		response.Error(c, err)
		return
	}
	item, err := h.service.Create(c.Request.Context(), auth.CurrentUser(c), warehouseID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, item)
}

// PUT /api/v1/warehouses/:id/locations/:location_id
func (h *WarehouseLocationHandler) Update(c *gin.Context) {
	warehouseID, locationID, err := locationIDs(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.UpdateWarehouseLocationRequest
	if err := bindJSON(c, &req); err != nil {
		response.Error(c, err)
		return
	}
	item, err := h.service.Update(c.Request.Context(), auth.CurrentUser(c), warehouseID, locationID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, item)
}

// DELETE /api/v1/warehouses/:id/locations/:location_id
func (h *WarehouseLocationHandler) Delete(c *gin.Context) {
	warehouseID, locationID, err := locationIDs(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.service.Delete(c.Request.Context(), auth.CurrentUser(c), warehouseID, locationID); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}
