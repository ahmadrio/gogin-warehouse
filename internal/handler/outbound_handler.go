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

type OutboundHandler struct {
	service *service.OutboundService
}

func NewOutboundHandler(s *service.OutboundService) *OutboundHandler {
	return &OutboundHandler{service: s}
}

// GET /api/v1/outbounds?warehouse_id=&status=&search=&page=
func (h *OutboundHandler) List(c *gin.Context) {
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

// GET /api/v1/outbounds/:id
func (h *OutboundHandler) Get(c *gin.Context) {
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

// POST /api/v1/outbounds
func (h *OutboundHandler) Create(c *gin.Context) {
	var req dto.OutboundRequest
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

// PUT /api/v1/outbounds/:id
func (h *OutboundHandler) Update(c *gin.Context) {
	id, err := paramUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.OutboundRequest
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

// POST /api/v1/outbounds/:id/approve
func (h *OutboundHandler) Approve(c *gin.Context) {
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

// POST /api/v1/outbounds/:id/reject
func (h *OutboundHandler) Reject(c *gin.Context) {
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
