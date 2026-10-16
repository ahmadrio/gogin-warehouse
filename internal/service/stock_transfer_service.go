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

// StockTransferService: alur sama dengan Out-Bound. Bedanya saat approve,
// SATU item menghasilkan DUA perubahan stok: keluar dari rak asal, masuk ke rak tujuan.
type StockTransferService struct {
	db         *gorm.DB
	transfers  *repository.StockTransferRepository
	warehouses *repository.WarehouseRepository
	lookups    *repository.LookupRepository
	sequences  *repository.SequenceRepository
	audits     *repository.AuditRepository
	stock      *StockService
}

func NewStockTransferService(db *gorm.DB, transfers *repository.StockTransferRepository, warehouses *repository.WarehouseRepository,
	lookups *repository.LookupRepository, sequences *repository.SequenceRepository, audits *repository.AuditRepository, stock *StockService) *StockTransferService {
	return &StockTransferService{db: db, transfers: transfers, warehouses: warehouses, lookups: lookups, sequences: sequences, audits: audits, stock: stock}
}

func (s *StockTransferService) List(ctx context.Context, user *auth.User, f repository.TransactionFilter, p pagination.Params) ([]model.StockTransfer, int64, error) {
	return s.transfers.List(ctx, user.WarehouseScope(), f, p)
}

func (s *StockTransferService) Get(ctx context.Context, user *auth.User, id uuid.UUID) (*model.StockTransfer, error) {
	h, err := s.transfers.FindByID(ctx, user.WarehouseScope(), id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.NotFound("stock transfer")
	}
	return h, err
}

func (s *StockTransferService) Create(ctx context.Context, user *auth.User, req dto.StockTransferRequest) (*model.StockTransfer, error) {
	if err := checkWarehouseAccess(ctx, s.warehouses, user, req.WarehouseID); err != nil {
		return nil, err
	}
	items, err := s.buildItems(ctx, req.WarehouseID, req)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	h := &model.StockTransfer{
		WarehouseID:     req.WarehouseID,
		Status:          model.StatusPending,
		Notes:           req.Notes,
		SubmittedAt:     &now,
		CreatedByUserID: user.ID,
		Items:           items,
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ref, err := s.sequences.WithTx(tx).Next(ctx, "TRF", now)
		if err != nil {
			return err
		}
		h.ReferenceNumber = ref
		if err := s.transfers.WithTx(tx).Create(ctx, h); err != nil {
			return err
		}
		return s.audits.WithTx(tx).Create(ctx, newAudit(user, model.ActionRequestApproval, model.EntityStockTransfer,
			h.ID, "membuat dan mengajukan "+h.ReferenceNumber))
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, user, h.ID)
}

func (s *StockTransferService) Update(ctx context.Context, user *auth.User, id uuid.UUID, req dto.StockTransferRequest) (*model.StockTransfer, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		h, err := s.lock(ctx, tx, user, id)
		if err != nil {
			return err
		}
		if h.Status == model.StatusApproved {
			return apperror.Conflict("stock transfer yang sudah approved tidak bisa diubah (BR-009)")
		}
		if h.CreatedByUserID != user.ID {
			return apperror.Forbidden("hanya pembuat pengajuan yang bisa merevisi")
		}
		items, err := s.buildItems(ctx, h.WarehouseID, req)
		if err != nil {
			return err
		}
		repo := s.transfers.WithTx(tx)
		if err := repo.ReplaceItems(ctx, h.ID, items); err != nil {
			return err
		}
		if err := repo.UpdateHeader(ctx, h.ID, map[string]any{
			"notes": req.Notes, "status": model.StatusPending, "submitted_at": time.Now(),
			"rejection_note": nil, "reviewed_at": nil, "reviewed_by_user_id": nil,
		}); err != nil {
			return err
		}
		return s.audits.WithTx(tx).Create(ctx, newAudit(user, model.ActionRequestApproval, model.EntityStockTransfer,
			h.ID, "merevisi dan mengajukan ulang "+h.ReferenceNumber))
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, user, id)
}

// Approve: setiap item = keluar dari rak asal + masuk ke rak tujuan, dalam SATU transaksi.
// Total stok produk di gudang tidak berubah (BR-007).
func (s *StockTransferService) Approve(ctx context.Context, user *auth.User, id uuid.UUID) (*model.StockTransfer, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		h, err := s.lock(ctx, tx, user, id)
		if err != nil {
			return err
		}
		if h.Status != model.StatusPending {
			return apperror.Conflict(fmt.Sprintf("hanya stock transfer pending yang bisa di-approve (status: %s)", h.Status))
		}
		items, err := s.transfers.WithTx(tx).ListItems(ctx, h.ID)
		if err != nil {
			return err
		}
		changes := make([]StockChange, 0, len(items)*2)
		for _, it := range items {
			itemID := it.ID
			changes = append(changes,
				StockChange{
					WarehouseID: h.WarehouseID, LocationID: it.SourceWarehouseLocationID, ProductID: it.ProductID,
					Direction: model.DirectionOut, Quantity: it.Quantity,
					MovementType: model.MovementStockTransfer, SourceID: h.ID, SourceItemID: &itemID,
					Label: fmt.Sprintf("%s di %s", it.Product.SKU, it.SourceLocation.Code),
				},
				StockChange{
					WarehouseID: h.WarehouseID, LocationID: it.DestinationWarehouseLocationID, ProductID: it.ProductID,
					Direction: model.DirectionIn, Quantity: it.Quantity,
					MovementType: model.MovementStockTransfer, SourceID: h.ID, SourceItemID: &itemID,
					Label: fmt.Sprintf("%s di %s", it.Product.SKU, it.DestinationLocation.Code),
				},
			)
		}
		if err := s.stock.Apply(ctx, tx, user, changes); err != nil {
			return err
		}
		if err := s.transfers.WithTx(tx).UpdateHeader(ctx, h.ID, map[string]any{
			"status": model.StatusApproved, "reviewed_at": time.Now(), "reviewed_by_user_id": user.ID,
		}); err != nil {
			return err
		}
		return s.audits.WithTx(tx).Create(ctx, newAudit(user, model.ActionApprove, model.EntityStockTransfer,
			h.ID, "menyetujui "+h.ReferenceNumber))
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, user, id)
}

func (s *StockTransferService) Reject(ctx context.Context, user *auth.User, id uuid.UUID, note string) (*model.StockTransfer, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		h, err := s.lock(ctx, tx, user, id)
		if err != nil {
			return err
		}
		if h.Status != model.StatusPending {
			return apperror.Conflict(fmt.Sprintf("hanya stock transfer pending yang bisa di-reject (status: %s)", h.Status))
		}
		if err := s.transfers.WithTx(tx).UpdateHeader(ctx, h.ID, map[string]any{
			"status": model.StatusRejected, "rejection_note": note,
			"reviewed_at": time.Now(), "reviewed_by_user_id": user.ID,
		}); err != nil {
			return err
		}
		return s.audits.WithTx(tx).Create(ctx, newAudit(user, model.ActionReject, model.EntityStockTransfer, h.ID, note))
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, user, id)
}

func (s *StockTransferService) lock(ctx context.Context, tx *gorm.DB, user *auth.User, id uuid.UUID) (*model.StockTransfer, error) {
	h, err := s.transfers.WithTx(tx).FindByIDForUpdate(ctx, user.WarehouseScope(), id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.NotFound("stock transfer")
	}
	return h, err
}

func (s *StockTransferService) buildItems(ctx context.Context, warehouseID uuid.UUID, req dto.StockTransferRequest) ([]model.StockTransferItem, error) {
	var productIDs, locationIDs []uuid.UUID
	for _, it := range req.Items {
		productIDs = append(productIDs, it.ProductID)
		locationIDs = append(locationIDs, it.SourceWarehouseLocationID, it.DestinationWarehouseLocationID)
	}
	master, err := loadMasterData(ctx, s.lookups, warehouseID, productIDs, locationIDs)
	if err != nil {
		return nil, err
	}

	errs := map[string]string{}
	seen := map[string]bool{}
	items := make([]model.StockTransferItem, 0, len(req.Items))
	for i, it := range req.Items {
		field := fmt.Sprintf("items[%d]", i)
		master.checkProduct(errs, field+".product_id", it.ProductID)
		// Kedua rak harus milik gudang yang sama -> transfer hanya di dalam satu gudang.
		master.checkLocation(errs, field+".source_warehouse_location_id", it.SourceWarehouseLocationID)
		master.checkLocation(errs, field+".destination_warehouse_location_id", it.DestinationWarehouseLocationID)
		if it.SourceWarehouseLocationID == it.DestinationWarehouseLocationID {
			errs[field+".destination_warehouse_location_id"] = "rak tujuan harus berbeda dengan rak asal"
		}
		key := it.ProductID.String() + "|" + it.SourceWarehouseLocationID.String() + "|" + it.DestinationWarehouseLocationID.String()
		if seen[key] {
			errs[field] = "item yang sama sudah ada"
		}
		seen[key] = true

		items = append(items, model.StockTransferItem{
			ProductID:                      it.ProductID,
			SourceWarehouseLocationID:      it.SourceWarehouseLocationID,
			DestinationWarehouseLocationID: it.DestinationWarehouseLocationID,
			Quantity:                       it.Quantity,
			Notes:                          it.Notes,
		})
	}
	if len(errs) > 0 {
		return nil, apperror.Validation(errs)
	}
	return items, nil
}
