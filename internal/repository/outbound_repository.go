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

// OutboundRepository: polanya sama persis dengan StockAdjustmentRepository (Tahap 13).
type OutboundRepository struct {
	db *gorm.DB
}

func NewOutboundRepository(db *gorm.DB) *OutboundRepository {
	return &OutboundRepository{db: db}
}

func (r *OutboundRepository) WithTx(tx *gorm.DB) *OutboundRepository {
	return &OutboundRepository{db: tx}
}

// TransactionFilter dipakai list In-Bound, Out-Bound, dan Stock Transfer.
type TransactionFilter struct {
	WarehouseID *uuid.UUID
	Status      string
}

func (r *OutboundRepository) List(ctx context.Context, scope auth.WarehouseScope, f TransactionFilter, p pagination.Params) ([]model.Outbound, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Outbound{}).
		Scopes(warehouseScope("outbounds.warehouse_id", scope))
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
	items := []model.Outbound{}
	err := q.Scopes(paginate(p)).Order("created_at DESC").Find(&items).Error
	return items, total, err
}

func (r *OutboundRepository) FindByID(ctx context.Context, scope auth.WarehouseScope, id uuid.UUID) (*model.Outbound, error) {
	var h model.Outbound
	err := r.db.WithContext(ctx).
		Scopes(warehouseScope("outbounds.warehouse_id", scope)).
		Preload("Items", func(db *gorm.DB) *gorm.DB { return db.Order("created_at ASC") }).
		Preload("Items.Product").
		Preload("Items.WarehouseLocation").
		First(&h, "id = ?", id).Error
	return &h, notFound(err)
}

// FindByIDForUpdate: SELECT ... FOR UPDATE pada header (approve/reject/revisi).
func (r *OutboundRepository) FindByIDForUpdate(ctx context.Context, scope auth.WarehouseScope, id uuid.UUID) (*model.Outbound, error) {
	var h model.Outbound
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Scopes(warehouseScope("outbounds.warehouse_id", scope)).
		First(&h, "id = ?", id).Error
	return &h, notFound(err)
}

func (r *OutboundRepository) ListItems(ctx context.Context, headerID uuid.UUID) ([]model.OutboundItem, error) {
	items := []model.OutboundItem{}
	err := r.db.WithContext(ctx).
		Preload("Product").
		Preload("WarehouseLocation").
		Where("outbound_id = ?", headerID).
		Find(&items).Error
	return items, err
}

// Create menyimpan header + item sekaligus (GORM ikut meng-INSERT h.Items).
func (r *OutboundRepository) Create(ctx context.Context, h *model.Outbound) error {
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *OutboundRepository) UpdateHeader(ctx context.Context, id uuid.UUID, fields map[string]any) error {
	return r.db.WithContext(ctx).Model(&model.Outbound{}).Where("id = ?", id).Updates(fields).Error
}

func (r *OutboundRepository) ReplaceItems(ctx context.Context, headerID uuid.UUID, items []model.OutboundItem) error {
	db := r.db.WithContext(ctx)
	if err := db.Where("outbound_id = ?", headerID).Delete(&model.OutboundItem{}).Error; err != nil {
		return err
	}
	for i := range items {
		items[i].OutboundID = headerID
	}
	return db.Create(&items).Error
}
