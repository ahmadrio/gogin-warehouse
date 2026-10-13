package auth

const (
	OpCreate  = "create"
	OpRead    = "read"
	OpUpdate  = "update"
	OpDelete  = "delete"
	OpApprove = "approve"
	OpReject  = "reject"
	OpExport  = "export"
)

const (
	ResUsers              = "users"
	ResProductCategories  = "product_categories"
	ResProducts           = "products"
	ResWarehouses         = "warehouses"
	ResWarehouseLocations = "warehouse_locations"
	ResInbounds           = "inbounds"
	ResOutbounds          = "outbounds"
	ResStockTransfers     = "stock_transfers"
	ResStockAdjustments   = "stock_adjustments"
	ResReports            = "reports"
	ResDashboard          = "dashboard"
	ResAuditTrails        = "audit_trails"
)

func Perm(resource, operation string) string {
	return resource + ":" + operation
}

const RoleSuperadmin = "superadmin"

var AllOperations = []string{
	OpCreate,
	OpRead,
	OpUpdate,
	OpDelete,
	OpApprove,
	OpReject,
	OpExport,
}

var AllResources = []string{
	ResUsers,
	ResProductCategories,
	ResProducts,
	ResWarehouses,
	ResWarehouseLocations,
	ResInbounds,
	ResOutbounds,
	ResStockTransfers,
	ResStockAdjustments,
	ResReports,
	ResDashboard,
	ResAuditTrails,
}

var RoleDefinisions = map[string][]string{
	"supervisor": {
		Perm(ResInbounds, OpRead), Perm(ResInbounds, OpApprove), Perm(ResInbounds, OpReject),
		Perm(ResOutbounds, OpRead), Perm(ResOutbounds, OpApprove), Perm(ResOutbounds, OpReject),
		Perm(ResStockTransfers, OpRead), Perm(ResStockTransfers, OpApprove), Perm(ResStockTransfers, OpReject),
		Perm(ResStockAdjustments, OpRead), Perm(ResStockAdjustments, OpApprove), Perm(ResStockAdjustments, OpReject),
		Perm(ResReports, OpRead), Perm(ResReports, OpExport), Perm(ResDashboard, OpRead),
		Perm(ResProductCategories, OpRead), Perm(ResProducts, OpRead),
		Perm(ResWarehouses, OpRead), Perm(ResWarehouseLocations, OpRead),
	},
	"staff": {
		Perm(ResInbounds, OpCreate), Perm(ResInbounds, OpRead), Perm(ResInbounds, OpUpdate),
		Perm(ResOutbounds, OpCreate), Perm(ResOutbounds, OpRead), Perm(ResOutbounds, OpUpdate),
		Perm(ResStockTransfers, OpCreate), Perm(ResStockTransfers, OpRead), Perm(ResStockTransfers, OpUpdate),
		Perm(ResStockAdjustments, OpCreate), Perm(ResStockAdjustments, OpRead), Perm(ResStockAdjustments, OpUpdate),
		Perm(ResDashboard, OpRead),
		Perm(ResProductCategories, OpRead), Perm(ResProducts, OpRead),
		Perm(ResWarehouses, OpRead), Perm(ResWarehouseLocations, OpRead),
	},
}
