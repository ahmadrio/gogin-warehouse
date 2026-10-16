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

// StockTransferRepository: polanya sama persis dengan StockAdjustmentRepository (Tahap 13).
type StockTransferRepository struct {
	db *gorm.DB
}

func NewStockTransferRepository(db *gorm.DB) *StockTransferRepository {
	return &StockTransferRepository{db: db}
}

func (r *StockTransferRepository) WithTx(tx *gorm.DB) *StockTransferRepository {
	return &StockTransferRepository{db: tx}
}

func (r *StockTransferRepository) List(ctx context.Context, scope auth.WarehouseScope, f TransactionFilter, p pagination.Params) ([]model.StockTransfer, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.StockTransfer{}).
		Scopes(warehouseScope("stock_transfers.warehouse_id", scope))
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
	items := []model.StockTransfer{}
	err := q.Scopes(paginate(p)).Order("created_at DESC").Find(&items).Error
	return items, total, err
}

func (r *StockTransferRepository) FindByID(ctx context.Context, scope auth.WarehouseScope, id uuid.UUID) (*model.StockTransfer, error) {
	var h model.StockTransfer
	err := r.db.WithContext(ctx).
		Scopes(warehouseScope("stock_transfers.warehouse_id", scope)).
		Preload("Items", func(db *gorm.DB) *gorm.DB { return db.Order("created_at ASC") }).
		Preload("Items.Product").
		Preload("Items.SourceLocation").
		Preload("Items.DestinationLocation").
		First(&h, "id = ?", id).Error
	return &h, notFound(err)
}

// FindByIDForUpdate: SELECT ... FOR UPDATE pada header (approve/reject/revisi).
func (r *StockTransferRepository) FindByIDForUpdate(ctx context.Context, scope auth.WarehouseScope, id uuid.UUID) (*model.StockTransfer, error) {
	var h model.StockTransfer
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Scopes(warehouseScope("stock_transfers.warehouse_id", scope)).
		First(&h, "id = ?", id).Error
	return &h, notFound(err)
}

func (r *StockTransferRepository) ListItems(ctx context.Context, headerID uuid.UUID) ([]model.StockTransferItem, error) {
	items := []model.StockTransferItem{}
	err := r.db.WithContext(ctx).
		Preload("Product").
		Preload("SourceLocation").
		Preload("DestinationLocation").
		Where("stock_transfer_id = ?", headerID).
		Find(&items).Error
	return items, err
}

// Create menyimpan header + item sekaligus (GORM ikut meng-INSERT h.Items).
func (r *StockTransferRepository) Create(ctx context.Context, h *model.StockTransfer) error {
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *StockTransferRepository) UpdateHeader(ctx context.Context, id uuid.UUID, fields map[string]any) error {
	return r.db.WithContext(ctx).Model(&model.StockTransfer{}).Where("id = ?", id).Updates(fields).Error
}

func (r *StockTransferRepository) ReplaceItems(ctx context.Context, headerID uuid.UUID, items []model.StockTransferItem) error {
	db := r.db.WithContext(ctx)
	if err := db.Where("stock_transfer_id = ?", headerID).Delete(&model.StockTransferItem{}).Error; err != nil {
		return err
	}
	for i := range items {
		items[i].StockTransferID = headerID
	}
	return db.Create(&items).Error
}
