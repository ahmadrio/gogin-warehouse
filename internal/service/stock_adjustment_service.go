package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"warehouse-management/internal/auth"
	"warehouse-management/internal/dto"
	"warehouse-management/internal/model"
	"warehouse-management/internal/repository"
	"warehouse-management/pkg/apperror"
	"warehouse-management/pkg/pagination"
)

type StockAdjustmentService struct {
	db          *gorm.DB
	adjustments *repository.StockAdjustmentRepository
	warehouses  *repository.WarehouseRepository
	lookups     *repository.LookupRepository
	sequences   *repository.SequenceRepository
	audits      *repository.AuditRepository
	stock       *StockService
}

func NewStockAdjustmentService(
	db *gorm.DB,
	adjustments *repository.StockAdjustmentRepository,
	warehouses *repository.WarehouseRepository,
	lookups *repository.LookupRepository,
	sequences *repository.SequenceRepository,
	audits *repository.AuditRepository,
	stock *StockService,
) *StockAdjustmentService {
	return &StockAdjustmentService{
		db: db, adjustments: adjustments, warehouses: warehouses, lookups: lookups,
		sequences: sequences, audits: audits, stock: stock,
	}
}

// ---------------------------------------------------------------- baca

func (s *StockAdjustmentService) List(ctx context.Context, user *auth.User, f repository.StockAdjustmentFilter, p pagination.Params) ([]model.StockAdjustment, int64, error) {
	return s.adjustments.List(ctx, user.WarehouseScope(), f, p)
}

func (s *StockAdjustmentService) Get(ctx context.Context, user *auth.User, id uuid.UUID) (*model.StockAdjustment, error) {
	adj, err := s.adjustments.FindByID(ctx, user.WarehouseScope(), id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.NotFound("stock adjustment")
	}
	return adj, err
}

// ---------------------------------------------------------------- staff: buat & revisi

// Create: staff membuat pengajuan baru, langsung berstatus pending (menunggu supervisor).
func (s *StockAdjustmentService) Create(ctx context.Context, user *auth.User, req dto.StockAdjustmentRequest) (*model.StockAdjustment, error) {
	if req.WarehouseID == uuid.Nil {
		return nil, apperror.Validation(map[string]string{"warehouse_id": "wajib diisi"})
	}
	// Gudang harus milik user (BR-002). Gudang orang lain dijawab 404.
	if _, err := s.warehouses.FindByID(ctx, user.WarehouseScope(), req.WarehouseID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.NotFound("warehouse")
		}
		return nil, err
	}

	items, err := s.buildItems(ctx, req.WarehouseID, req)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	adj := &model.StockAdjustment{
		WarehouseID:     req.WarehouseID,
		Status:          model.StatusPending,
		AdjustmentType:  req.AdjustmentType,
		Reason:          req.Reason,
		SubmittedAt:     &now,
		CreatedByUserID: user.ID,
		Items:           items,
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ref, err := s.sequences.WithTx(tx).Next(ctx, "ADJ", now)
		if err != nil {
			return err
		}
		adj.ReferenceNumber = ref

		if err := s.adjustments.WithTx(tx).Create(ctx, adj); err != nil {
			return err
		}
		return s.audits.WithTx(tx).Create(ctx, newAudit(user, model.ActionRequestApproval, model.EntityStockAdjustment,
			adj.ID, fmt.Sprintf("membuat dan mengajukan %s", adj.ReferenceNumber)))
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, user, adj.ID)
}

// Update: staff merevisi pengajuan yang masih pending atau sudah ditolak,
// lalu otomatis diajukan ulang (status kembali pending).
func (s *StockAdjustmentService) Update(ctx context.Context, user *auth.User, id uuid.UUID, req dto.StockAdjustmentRequest) (*model.StockAdjustment, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		adj, err := s.lockForReview(ctx, tx, user, id)
		if err != nil {
			return err
		}
		if adj.Status == model.StatusApproved {
			return apperror.Conflict("pengajuan yang sudah approved tidak bisa diubah (BR-009)")
		}
		if adj.CreatedByUserID != user.ID {
			return apperror.Forbidden("hanya pembuat pengajuan yang bisa merevisi")
		}

		items, err := s.buildItems(ctx, adj.WarehouseID, req)
		if err != nil {
			return err
		}

		repo := s.adjustments.WithTx(tx)
		if err := repo.ReplaceItems(ctx, adj.ID, items); err != nil {
			return err
		}
		if err := repo.UpdateHeader(ctx, adj.ID, map[string]any{
			"adjustment_type":     req.AdjustmentType,
			"reason":              req.Reason,
			"status":              model.StatusPending,
			"submitted_at":        time.Now(),
			"rejection_note":      nil,
			"reviewed_at":         nil,
			"reviewed_by_user_id": nil,
		}); err != nil {
			return err
		}
		return s.audits.WithTx(tx).Create(ctx, newAudit(user, model.ActionRequestApproval, model.EntityStockAdjustment,
			adj.ID, fmt.Sprintf("merevisi dan mengajukan ulang %s", adj.ReferenceNumber)))
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, user, id)
}

// ---------------------------------------------------------------- supervisor: approve & reject

// Approve: stok berubah HANYA di sini (BR-004). Semua langkah dalam satu transaksi.
func (s *StockAdjustmentService) Approve(ctx context.Context, user *auth.User, id uuid.UUID) (*model.StockAdjustment, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Kunci header. Approve bersamaan akan antre di sini.
		adj, err := s.lockForReview(ctx, tx, user, id)
		if err != nil {
			return err
		}
		if adj.Status != model.StatusPending {
			return apperror.Conflict(fmt.Sprintf("hanya pengajuan pending yang bisa di-approve (status: %s)", adj.Status))
		}

		// 2. Ubah setiap item menjadi perubahan stok.
		items, err := s.adjustments.WithTx(tx).ListItems(ctx, adj.ID)
		if err != nil {
			return err
		}
		changes := make([]StockChange, 0, len(items))
		for _, it := range items {
			direction, qty := model.DirectionIn, it.QuantityChange
			if it.QuantityChange < 0 {
				direction, qty = model.DirectionOut, -it.QuantityChange
			}
			itemID := it.ID
			changes = append(changes, StockChange{
				WarehouseID:  adj.WarehouseID,
				LocationID:   it.WarehouseLocationID,
				ProductID:    it.ProductID,
				Direction:    direction,
				Quantity:     qty,
				MovementType: model.MovementStockAdjustment,
				SourceID:     adj.ID,
				SourceItemID: &itemID,
				Label:        fmt.Sprintf("%s di %s", it.Product.SKU, it.WarehouseLocation.Code),
			})
		}

		// 3. Ubah saldo + tulis stock movement (stok tidak cukup = error = rollback semua).
		if err := s.stock.Apply(ctx, tx, user, changes); err != nil {
			return err
		}

		// 4. Tandai approved.
		if err := s.adjustments.WithTx(tx).UpdateHeader(ctx, adj.ID, map[string]any{
			"status":              model.StatusApproved,
			"reviewed_at":         time.Now(),
			"reviewed_by_user_id": user.ID,
		}); err != nil {
			return err
		}
		return s.audits.WithTx(tx).Create(ctx, newAudit(user, model.ActionApprove, model.EntityStockAdjustment,
			adj.ID, fmt.Sprintf("menyetujui %s", adj.ReferenceNumber)))
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, user, id)
}

// Reject: tidak menyentuh stok, hanya mengubah status dan menyimpan alasan penolakan.
func (s *StockAdjustmentService) Reject(ctx context.Context, user *auth.User, id uuid.UUID, note string) (*model.StockAdjustment, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		adj, err := s.lockForReview(ctx, tx, user, id)
		if err != nil {
			return err
		}
		if adj.Status != model.StatusPending {
			return apperror.Conflict(fmt.Sprintf("hanya pengajuan pending yang bisa di-reject (status: %s)", adj.Status))
		}
		if err := s.adjustments.WithTx(tx).UpdateHeader(ctx, adj.ID, map[string]any{
			"status":              model.StatusRejected,
			"rejection_note":      note,
			"reviewed_at":         time.Now(),
			"reviewed_by_user_id": user.ID,
		}); err != nil {
			return err
		}
		// Deskripsi audit = catatan penolakan (sesuai requirement Audit Trail poin 6)
		return s.audits.WithTx(tx).Create(ctx, newAudit(user, model.ActionReject, model.EntityStockAdjustment,
			adj.ID, note))
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, user, id)
}

// ---------------------------------------------------------------- helper

// lockForReview mengambil + mengunci header di dalam transaksi, dengan scope gudang user.
func (s *StockAdjustmentService) lockForReview(ctx context.Context, tx *gorm.DB, user *auth.User, id uuid.UUID) (*model.StockAdjustment, error) {
	adj, err := s.adjustments.WithTx(tx).FindByIDForUpdate(ctx, user.WarehouseScope(), id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.NotFound("stock adjustment")
	}
	return adj, err
}

// buildItems memvalidasi item request terhadap data master, lalu mengubahnya menjadi model.
func (s *StockAdjustmentService) buildItems(ctx context.Context, warehouseID uuid.UUID, req dto.StockAdjustmentRequest) ([]model.StockAdjustmentItem, error) {
	productIDs := make([]uuid.UUID, 0, len(req.Items))
	locationIDs := make([]uuid.UUID, 0, len(req.Items))
	for _, it := range req.Items {
		productIDs = append(productIDs, it.ProductID)
		locationIDs = append(locationIDs, it.WarehouseLocationID)
	}

	products, err := s.lookups.ActiveProducts(ctx, productIDs)
	if err != nil {
		return nil, err
	}
	locations, err := s.lookups.ActiveLocations(ctx, warehouseID, locationIDs)
	if err != nil {
		return nil, err
	}

	errs := map[string]string{}
	seen := map[string]bool{}
	items := make([]model.StockAdjustmentItem, 0, len(req.Items))

	for i, it := range req.Items {
		field := fmt.Sprintf("items[%d]", i)

		if _, ok := products[it.ProductID]; !ok {
			errs[field+".product_id"] = "produk tidak ditemukan atau tidak aktif"
		}
		if _, ok := locations[it.WarehouseLocationID]; !ok {
			errs[field+".warehouse_location_id"] = "lokasi tidak ditemukan, tidak aktif, atau bukan milik warehouse ini"
		}

		key := it.WarehouseLocationID.String() + "|" + it.ProductID.String()
		if seen[key] {
			errs[field] = "produk yang sama di lokasi yang sama sudah ada di item lain"
		}
		seen[key] = true

		// Barang rusak dan hilang hanya boleh MENGURANGI stok.
		reduceOnly := req.AdjustmentType == model.AdjustmentDamagedGoods || req.AdjustmentType == model.AdjustmentLostGoods
		if reduceOnly && it.QuantityChange > 0 {
			errs[field+".quantity_change"] = "harus negatif untuk tipe " + req.AdjustmentType
		}

		items = append(items, model.StockAdjustmentItem{
			ProductID:           it.ProductID,
			WarehouseLocationID: it.WarehouseLocationID,
			QuantityChange:      it.QuantityChange,
			Notes:               it.Notes,
		})
	}

	if len(errs) > 0 {
		return nil, apperror.Validation(errs)
	}
	return items, nil
}
