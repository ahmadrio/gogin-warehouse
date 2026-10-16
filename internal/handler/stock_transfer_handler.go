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

type StockTransferHandler struct {
	service *service.StockTransferService
}

func NewStockTransferHandler(s *service.StockTransferService) *StockTransferHandler {
	return &StockTransferHandler{service: s}
}

// GET /api/v1/stock-transfers?warehouse_id=&status=&search=&page=
func (h *StockTransferHandler) List(c *gin.Context) {
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

// GET /api/v1/stock-transfers/:id
func (h *StockTransferHandler) Get(c *gin.Context) {
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

// POST /api/v1/stock-transfers
func (h *StockTransferHandler) Create(c *gin.Context) {
	var req dto.StockTransferRequest
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

// PUT /api/v1/stock-transfers/:id
func (h *StockTransferHandler) Update(c *gin.Context) {
	id, err := paramUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.StockTransferRequest
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

// POST /api/v1/stock-transfers/:id/approve
func (h *StockTransferHandler) Approve(c *gin.Context) {
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

// POST /api/v1/stock-transfers/:id/reject
func (h *StockTransferHandler) Reject(c *gin.Context) {
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
