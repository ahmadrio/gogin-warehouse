package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"warehouse-management/internal/model"
)

type WarehouseLocationRepository struct {
	db *gorm.DB
}

func NewWarehouseLocationRepository(db *gorm.DB) *WarehouseLocationRepository {
	return &WarehouseLocationRepository{db: db}
}

func (r *WarehouseLocationRepository) WithTx(tx *gorm.DB) *WarehouseLocationRepository {
	return &WarehouseLocationRepository{db: tx}
}

// FindByID mencari lokasi DI DALAM gudang tertentu. Lokasi milik gudang lain = tidak ditemukan.
func (r *WarehouseLocationRepository) FindByID(ctx context.Context, warehouseID, id uuid.UUID) (*model.WarehouseLocation, error) {
	var l model.WarehouseLocation
	err := r.db.WithContext(ctx).First(&l, "id = ? AND warehouse_id = ?", id, warehouseID).Error
	return &l, notFound(err)
}

// ExistsByCode: kode lokasi unik PER GUDANG (RAK-A-01 boleh ada di Malang dan Surabaya).
func (r *WarehouseLocationRepository) ExistsByCode(ctx context.Context, warehouseID uuid.UUID, code string, excludeID *uuid.UUID) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.WarehouseLocation{}).
		Where("warehouse_id = ? AND code = ?", warehouseID, code)
	if excludeID != nil {
		q = q.Where("id <> ?", *excludeID)
	}
	var count int64
	err := q.Count(&count).Error
	return count > 0, err
}

// TotalStock: jumlah semua barang yang masih ada di lokasi ini.
func (r *WarehouseLocationRepository) TotalStock(ctx context.Context, id uuid.UUID) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&model.WarehouseLocationItem{}).
		Where("warehouse_location_id = ?", id).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&total).Error
	return total, err
}

func (r *WarehouseLocationRepository) Create(ctx context.Context, l *model.WarehouseLocation) error {
	return r.db.WithContext(ctx).Create(l).Error
}

func (r *WarehouseLocationRepository) Update(ctx context.Context, l *model.WarehouseLocation) error {
	return r.db.WithContext(ctx).Save(l).Error
}

func (r *WarehouseLocationRepository) Delete(ctx context.Context, l *model.WarehouseLocation) error {
	return r.db.WithContext(ctx).Delete(l).Error
}
