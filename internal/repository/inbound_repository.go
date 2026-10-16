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

// InboundRepository: polanya sama persis dengan StockAdjustmentRepository (Tahap 13),
// ditambah SetItemLocation untuk putaway.
type InboundRepository struct {
	db *gorm.DB
}

func NewInboundRepository(db *gorm.DB) *InboundRepository {
	return &InboundRepository{db: db}
}

func (r *InboundRepository) WithTx(tx *gorm.DB) *InboundRepository {
	return &InboundRepository{db: tx}
}

func (r *InboundRepository) List(ctx context.Context, scope auth.WarehouseScope, f TransactionFilter, p pagination.Params) ([]model.Inbound, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Inbound{}).
		Scopes(warehouseScope("inbounds.warehouse_id", scope))
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
	items := []model.Inbound{}
	err := q.Scopes(paginate(p)).Order("created_at DESC").Find(&items).Error
	return items, total, err
}

func (r *InboundRepository) FindByID(ctx context.Context, scope auth.WarehouseScope, id uuid.UUID) (*model.Inbound, error) {
	var h model.Inbound
	err := r.db.WithContext(ctx).
		Scopes(warehouseScope("inbounds.warehouse_id", scope)).
		Preload("Items", func(db *gorm.DB) *gorm.DB { return db.Order("created_at ASC") }).
		Preload("Items.Product").
		Preload("Items.WarehouseLocation").
		First(&h, "id = ?", id).Error
	return &h, notFound(err)
}

// FindByIDForUpdate: SELECT ... FOR UPDATE pada header (approve/reject/revisi/putaway).
func (r *InboundRepository) FindByIDForUpdate(ctx context.Context, scope auth.WarehouseScope, id uuid.UUID) (*model.Inbound, error) {
	var h model.Inbound
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Scopes(warehouseScope("inbounds.warehouse_id", scope)).
		First(&h, "id = ?", id).Error
	return &h, notFound(err)
}

func (r *InboundRepository) ListItems(ctx context.Context, headerID uuid.UUID) ([]model.InboundItem, error) {
	items := []model.InboundItem{}
	err := r.db.WithContext(ctx).
		Preload("Product").
		Preload("WarehouseLocation").
		Where("inbound_id = ?", headerID).
		Find(&items).Error
	return items, err
}

// Create menyimpan header + item sekaligus (GORM ikut meng-INSERT h.Items).
func (r *InboundRepository) Create(ctx context.Context, h *model.Inbound) error {
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *InboundRepository) UpdateHeader(ctx context.Context, id uuid.UUID, fields map[string]any) error {
	return r.db.WithContext(ctx).Model(&model.Inbound{}).Where("id = ?", id).Updates(fields).Error
}

func (r *InboundRepository) ReplaceItems(ctx context.Context, headerID uuid.UUID, items []model.InboundItem) error {
	db := r.db.WithContext(ctx)
	if err := db.Where("inbound_id = ?", headerID).Delete(&model.InboundItem{}).Error; err != nil {
		return err
	}
	for i := range items {
		items[i].InboundID = headerID
	}
	return db.Create(&items).Error
}

// SetItemLocation mengisi rak tujuan satu item saat putaway.
func (r *InboundRepository) SetItemLocation(ctx context.Context, itemID, locationID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.InboundItem{}).
		Where("id = ?", itemID).
		Update("warehouse_location_id", locationID).Error
}
