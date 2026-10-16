package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"warehouse-management/internal/model"
)

// LookupRepository mengambil data master yang dibutuhkan transaksi untuk validasi.
type LookupRepository struct {
	db *gorm.DB
}

func NewLookupRepository(db *gorm.DB) *LookupRepository {
	return &LookupRepository{db: db}
}

// ActiveProducts mengembalikan map id -> produk, hanya untuk produk yang aktif.
func (r *LookupRepository) ActiveProducts(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]model.Product, error) {
	var products []model.Product
	if err := r.db.WithContext(ctx).Where("id IN ? AND is_active = true", ids).Find(&products).Error; err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]model.Product, len(products))
	for _, p := range products {
		result[p.ID] = p
	}
	return result, nil
}

// ActiveLocations mengembalikan map id -> lokasi, hanya lokasi aktif di gudang warehouseID.
func (r *LookupRepository) ActiveLocations(ctx context.Context, warehouseID uuid.UUID, ids []uuid.UUID) (map[uuid.UUID]model.WarehouseLocation, error) {
	var locations []model.WarehouseLocation
	if err := r.db.WithContext(ctx).
		Where("id IN ? AND warehouse_id = ? AND is_active = true", ids, warehouseID).
		Find(&locations).Error; err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]model.WarehouseLocation, len(locations))
	for _, l := range locations {
		result[l.ID] = l
	}
	return result, nil
}
