package model

// Nilai kolom audit_trails.action
const (
	ActionCreate          = "create"
	ActionUpdate          = "update"
	ActionDelete          = "delete"
	ActionRequestApproval = "request_approval"
	ActionApprove         = "approve"
	ActionReject          = "reject"
	ActionExport          = "export"
)

// Nilai kolom audit_trails.entity_type
const (
	EntityProductCategory   = "product_category"
	EntityProduct           = "product"
	EntityWarehouse         = "warehouse"
	EntityWarehouseLocation = "warehouse_location"
	EntityInbound           = "inbound"
	EntityOutbound          = "outbound"
	EntityStockTransfer     = "stock_transfer"
	EntityStockAdjustment   = "stock_adjustment"
)
