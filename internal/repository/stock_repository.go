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

type StockRepository struct {
	db *gorm.DB
}

func NewStockRepository(db *gorm.DB) *StockRepository {
	return &StockRepository{db: db}
}

func (r *StockRepository) WithTx(tx *gorm.DB) *StockRepository {
	return &StockRepository{db: tx}
}

// LockBalance mengambil baris saldo (lokasi, produk) dan MENGUNCINYA sampai transaksi selesai.
// Kalau barisnya belum ada, dibuat dulu dengan quantity 0.
// Wajib dipanggil di dalam transaksi (lewat WithTx), kalau tidak kuncinya langsung lepas.
func (r *StockRepository) LockBalance(ctx context.Context, warehouseID, locationID, productID uuid.UUID) (*model.WarehouseLocationItem, error) {
	db := r.db.WithContext(ctx)

	// Langkah 1: pastikan baris ada (aman walau dua request bersamaan).
	empty := model.WarehouseLocationItem{
		WarehouseID:         warehouseID,
		WarehouseLocationID: locationID,
		ProductID:           productID,
		Quantity:            0,
	}
	err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "warehouse_location_id"}, {Name: "product_id"}},
		DoNothing: true,
	}).Create(&empty).Error
	if err != nil {
		return nil, err
	}

	// Langkah 2: SELECT ... FOR UPDATE -> kunci baris.
	var balance model.WarehouseLocationItem
	err = db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("warehouse_location_id = ? AND product_id = ?", locationID, productID).
		First(&balance).Error
	return &balance, err
}

func (r *StockRepository) UpdateQuantity(ctx context.Context, balance *model.WarehouseLocationItem) error {
	return r.db.WithContext(ctx).Model(balance).Update("quantity", balance.Quantity).Error
}

func (r *StockRepository) CreateMovement(ctx context.Context, m *model.StockMovement) error {
	return r.db.WithContext(ctx).Create(m).Error
}

// ---- Query saldo untuk ditampilkan ----

type StockFilter struct {
	WarehouseID *uuid.UUID
	LocationID  *uuid.UUID
	ProductID   *uuid.UUID
}

// StockBalanceRow adalah hasil JOIN saldo + produk + lokasi, siap ditampilkan.
type StockBalanceRow struct {
	WarehouseID         uuid.UUID `json:"warehouse_id"`
	WarehouseLocationID uuid.UUID `json:"warehouse_location_id"`
	LocationCode        string    `json:"location_code"`
	ProductID           uuid.UUID `json:"product_id"`
	SKU                 string    `gorm:"column:sku" json:"sku"`
	ProductName         string    `json:"product_name"`
	Unit                string    `json:"unit"`
	Quantity            int       `json:"quantity"`
}

func (r *StockRepository) ListBalances(ctx context.Context, scope auth.WarehouseScope, f StockFilter, p pagination.Params) ([]StockBalanceRow, int64, error) {
	q := r.db.WithContext(ctx).
		Table("warehouse_location_items wli").
		Joins("JOIN products p ON p.id = wli.product_id").
		Joins("JOIN warehouse_locations wl ON wl.id = wli.warehouse_location_id").
		Scopes(warehouseScope("wli.warehouse_id", scope)).
		Where("wli.quantity > 0")

	if f.WarehouseID != nil {
		q = q.Where("wli.warehouse_id = ?", *f.WarehouseID)
	}
	if f.LocationID != nil {
		q = q.Where("wli.warehouse_location_id = ?", *f.LocationID)
	}
	if f.ProductID != nil {
		q = q.Where("wli.product_id = ?", *f.ProductID)
	}
	if p.Search != "" {
		like := "%" + p.Search + "%"
		q = q.Where("p.sku ILIKE ? OR p.name ILIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := []StockBalanceRow{}
	err := q.Select(`wli.warehouse_id, wli.warehouse_location_id, wl.code AS location_code,
			wli.product_id, p.sku, p.name AS product_name, p.unit, wli.quantity`).
		Scopes(paginate(p)).
		Order("p.sku ASC, wl.code ASC").
		Scan(&rows).Error
	return rows, total, err
}
