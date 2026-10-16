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

type InboundHandler struct {
	service *service.InboundService
}

func NewInboundHandler(s *service.InboundService) *InboundHandler {
	return &InboundHandler{service: s}
}

// GET /api/v1/inbounds?warehouse_id=&status=&search=&page=
func (h *InboundHandler) List(c *gin.Context) {
	warehouseID, err := queryUUID(c, "warehouse_id")
	if err != nil {
		response.Error(c, err)
		return
	}
	f := repository.TransactionFilter{WarehouseID: warehouseID, Status: c.Query("status")}

	p := pagination.FromQuery(c)
	items, total, err := h.service.List(c.Request.Context(), auth.CurrentUser(c), f, p)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, items, pagination.NewMeta(p, total))
}

// GET /api/v1/inbounds/:id
func (h *InboundHandler) Get(c *gin.Context) {
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

// POST /api/v1/inbounds
func (h *InboundHandler) Create(c *gin.Context) {
	var req dto.InboundRequest
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

// PUT /api/v1/inbounds/:id
func (h *InboundHandler) Update(c *gin.Context) {
	id, err := paramUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.InboundRequest
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

// POST /api/v1/inbounds/:id/approve
func (h *InboundHandler) Approve(c *gin.Context) {
	id, err := paramUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	item, err := h.service.Approve(c.Request.Context(), auth.CurrentUser(c), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, item)
}

// POST /api/v1/inbounds/:id/reject
func (h *InboundHandler) Reject(c *gin.Context) {
	id, err := paramUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.RejectRequest
	if err := bindJSON(c, &req); err != nil {
		response.Error(c, err)
		return
	}
	item, err := h.service.Reject(c.Request.Context(), auth.CurrentUser(c), id, req.RejectionNote)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, item)
}

// POST /api/v1/inbounds/:id/putaway
func (h *InboundHandler) Putaway(c *gin.Context) {
	id, err := paramUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.PutawayRequest
	if err := bindJSON(c, &req); err != nil {
		response.Error(c, err)
		return
	}
	item, err := h.service.Putaway(c.Request.Context(), auth.CurrentUser(c), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, item)
}
