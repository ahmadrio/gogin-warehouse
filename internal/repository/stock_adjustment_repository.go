package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"warehouse-management/internal/auth"
	"warehouse-management/internal/model"
	"warehouse-management/pkg/pagination"
)

type StockAdjustmentRepository struct {
	db *gorm.DB
}

func NewStockAdjustmentRepository(db *gorm.DB) *StockAdjustmentRepository {
	return &StockAdjustmentRepository{db: db}
}

func (r *StockAdjustmentRepository) WithTx(tx *gorm.DB) *StockAdjustmentRepository {
	return &StockAdjustmentRepository{db: tx}
}

type StockAdjustmentFilter struct {
	WarehouseID *uuid.UUID
	Status      string
}

func (r *StockAdjustmentRepository) List(ctx context.Context, scope auth.WarehouseScope, f StockAdjustmentFilter, p pagination.Params) ([]model.StockAdjustment, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.StockAdjustment{}).
		Scopes(warehouseScope("stock_adjustments.warehouse_id", scope))
	if f.WarehouseID != nil {
		q = q.Where("warehouse_id = ?", *f.WarehouseID)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if p.Search != "" {
		q = q.Where("reference_number ILIKE ?", "%"+p.Search+"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := []model.StockAdjustment{}
	err := q.Scopes(paginate(p)).Order("created_at DESC").Find(&items).Error
	return items, total, err
}

// FindByID mengambil header beserta item, produk, dan lokasinya (untuk halaman detail).
func (r *StockAdjustmentRepository) FindByID(ctx context.Context, scope auth.WarehouseScope, id uuid.UUID) (*model.StockAdjustment, error) {
	var adj model.StockAdjustment
	err := r.db.WithContext(ctx).
		Scopes(warehouseScope("stock_adjustments.warehouse_id", scope)).
		Preload("Items", func(db *gorm.DB) *gorm.DB { return db.Order("created_at ASC") }).
		Preload("Items.Product").
		Preload("Items.WarehouseLocation").
		First(&adj, "id = ?", id).Error
	return &adj, notFound(err)
}

// FindByIDForUpdate mengambil header DAN menguncinya (SELECT ... FOR UPDATE).
// Kalau dua supervisor menekan approve bersamaan, yang kedua menunggu,
// lalu melihat status sudah "approved" dan ditolak.
func (r *StockAdjustmentRepository) FindByIDForUpdate(ctx context.Context, scope auth.WarehouseScope, id uuid.UUID) (*model.StockAdjustment, error) {
	var adj model.StockAdjustment
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Scopes(warehouseScope("stock_adjustments.warehouse_id", scope)).
		First(&adj, "id = ?", id).Error
	return &adj, notFound(err)
}

func (r *StockAdjustmentRepository) ListItems(ctx context.Context, adjustmentID uuid.UUID) ([]model.StockAdjustmentItem, error) {
	items := []model.StockAdjustmentItem{}
	err := r.db.WithContext(ctx).
		Preload("Product").
		Preload("WarehouseLocation").
		Where("stock_adjustment_id = ?", adjustmentID).
		Find(&items).Error
	return items, err
}

// Create menyimpan header. Karena adj.Items terisi, GORM otomatis ikut
// meng-INSERT semua item dan mengisi stock_adjustment_id-nya.
func (r *StockAdjustmentRepository) Create(ctx context.Context, adj *model.StockAdjustment) error {
	return r.db.WithContext(ctx).Create(adj).Error
}

// UpdateHeader mengubah kolom tertentu. Pakai map agar nilai nil/kosong ikut tersimpan.
func (r *StockAdjustmentRepository) UpdateHeader(ctx context.Context, id uuid.UUID, fields map[string]any) error {
	return r.db.WithContext(ctx).Model(&model.StockAdjustment{}).Where("id = ?", id).Updates(fields).Error
}

// ReplaceItems menghapus semua item lama lalu menyimpan item baru (dipakai saat revisi).
func (r *StockAdjustmentRepository) ReplaceItems(ctx context.Context, adjustmentID uuid.UUID, items []model.StockAdjustmentItem) error {
	db := r.db.WithContext(ctx)
	if err := db.Where("stock_adjustment_id = ?", adjustmentID).Delete(&model.StockAdjustmentItem{}).Error; err != nil {
		return err
	}
	for i := range items {
		items[i].StockAdjustmentID = adjustmentID
	}
	return db.Create(&items).Error
}
