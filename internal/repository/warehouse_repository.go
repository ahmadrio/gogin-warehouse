package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"warehouse-management/internal/auth"
	"warehouse-management/internal/model"
	"warehouse-management/pkg/pagination"
)

type WarehouseRepository struct {
	db *gorm.DB
}

func NewWarehouseRepository(db *gorm.DB) *WarehouseRepository {
	return &WarehouseRepository{db: db}
}

func (r *WarehouseRepository) WithTx(tx *gorm.DB) *WarehouseRepository {
	return &WarehouseRepository{db: tx}
}

// List hanya mengembalikan gudang yang boleh diakses (scope).
func (r *WarehouseRepository) List(ctx context.Context, scope auth.WarehouseScope, p pagination.Params) ([]model.Warehouse, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Warehouse{}).
		Scopes(warehouseScope("warehouses.id", scope))
	if p.Search != "" {
		like := "%" + p.Search + "%"
		q = q.Where("code ILIKE ? OR name ILIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := []model.Warehouse{}
	err := q.Scopes(paginate(p)).Order("code ASC").Find(&items).Error
	return items, total, err
}

// FindByID juga memakai scope: gudang milik orang lain dianggap "tidak ada" (404).
func (r *WarehouseRepository) FindByID(ctx context.Context, scope auth.WarehouseScope, id uuid.UUID) (*model.Warehouse, error) {
	var w model.Warehouse
	err := r.db.WithContext(ctx).
		Scopes(warehouseScope("warehouses.id", scope)).
		First(&w, "id = ?", id).Error
	return &w, notFound(err)
}

func (r *WarehouseRepository) ListLocations(ctx context.Context, warehouseID uuid.UUID) ([]model.WarehouseLocation, error) {
	items := []model.WarehouseLocation{}
	err := r.db.WithContext(ctx).Where("warehouse_id = ?", warehouseID).Order("code ASC").Find(&items).Error
	return items, err
}

func (r *WarehouseRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Warehouse{}).Where("code = ?", code).Count(&count).Error
	return count > 0, err
}

func (r *WarehouseRepository) Create(ctx context.Context, w *model.Warehouse) error {
	return r.db.WithContext(ctx).Create(w).Error
}
