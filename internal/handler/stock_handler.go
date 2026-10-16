package handler

import (
	"github.com/gin-gonic/gin"

	"warehouse-management/internal/auth"
	"warehouse-management/internal/repository"
	"warehouse-management/internal/service"
	"warehouse-management/pkg/pagination"
	"warehouse-management/pkg/response"
)

type StockHandler struct {
	service *service.StockService
}

func NewStockHandler(s *service.StockService) *StockHandler {
	return &StockHandler{service: s}
}

// GET /api/v1/stocks?warehouse_id=&location_id=&product_id=&search=&page=
func (h *StockHandler) List(c *gin.Context) {
	var f repository.StockFilter
	var err error
	if f.WarehouseID, err = queryUUID(c, "warehouse_id"); err != nil {
		response.Error(c, err)
		return
	}
	if f.LocationID, err = queryUUID(c, "location_id"); err != nil {
		response.Error(c, err)
		return
	}
	if f.ProductID, err = queryUUID(c, "product_id"); err != nil {
		response.Error(c, err)
		return
	}

	p := pagination.FromQuery(c)
	rows, total, err := h.service.ListBalances(c.Request.Context(), auth.CurrentUser(c), f, p)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, rows, pagination.NewMeta(p, total))
}
