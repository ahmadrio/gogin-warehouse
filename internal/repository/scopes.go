package repository

import (
	"gorm.io/gorm"

	"warehouse-management/internal/auth"
	"warehouse-management/pkg/pagination"
)

// paginate menambah LIMIT & OFFSET.
func paginate(p pagination.Params) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Limit(p.PerPage).Offset(p.Offset())
	}
}

// warehouseScope membatasi query hanya ke gudang yang boleh diakses user.
// column: nama kolom warehouse id, mis. "warehouses.id" atau "inbounds.warehouse_id".
func warehouseScope(column string, scope auth.WarehouseScope) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if scope.All {
			return db
		}
		if len(scope.IDs) == 0 {
			return db.Where("1 = 0") // tidak punya gudang sama sekali -> tidak ada data
		}
		return db.Where(column+" IN ?", scope.IDs)
	}
}
