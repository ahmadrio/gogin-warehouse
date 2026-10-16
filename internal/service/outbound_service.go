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

// OutboundService: alurnya sama dengan StockAdjustmentService (Tahap 13).
// Bedanya hanya: semua item MENGURANGI stok, dan quantity selalu positif.
type OutboundService struct {
	db         *gorm.DB
	outbounds  *repository.OutboundRepository
	warehouses *repository.WarehouseRepository
	lookups    *repository.LookupRepository
	sequences  *repository.SequenceRepository
	audits     *repository.AuditRepository
	stock      *StockService
}

func NewOutboundService(db *gorm.DB, outbounds *repository.OutboundRepository, warehouses *repository.WarehouseRepository,
	lookups *repository.LookupRepository, sequences *repository.SequenceRepository, audits *repository.AuditRepository, stock *StockService) *OutboundService {
	return &OutboundService{db: db, outbounds: outbounds, warehouses: warehouses, lookups: lookups, sequences: sequences, audits: audits, stock: stock}
}

func (s *OutboundService) List(ctx context.Context, user *auth.User, f repository.TransactionFilter, p pagination.Params) ([]model.Outbound, int64, error) {
	return s.outbounds.List(ctx, user.WarehouseScope(), f, p)
}

func (s *OutboundService) Get(ctx context.Context, user *auth.User, id uuid.UUID) (*model.Outbound, error) {
	h, err := s.outbounds.FindByID(ctx, user.WarehouseScope(), id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.NotFound("out-bound")
	}
	return h, err
}

func (s *OutboundService) Create(ctx context.Context, user *auth.User, req dto.OutboundRequest) (*model.Outbound, error) {
	if err := checkWarehouseAccess(ctx, s.warehouses, user, req.WarehouseID); err != nil {
		return nil, err
	}
	items, err := s.buildItems(ctx, req.WarehouseID, req)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	h := &model.Outbound{
		WarehouseID:     req.WarehouseID,
		Status:          model.StatusPending,
		Notes:           req.Notes,
		SubmittedAt:     &now,
		CreatedByUserID: user.ID,
		Items:           items,
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ref, err := s.sequences.WithTx(tx).Next(ctx, "OUT", now)
		if err != nil {
			return err
		}
		h.ReferenceNumber = ref
		if err := s.outbounds.WithTx(tx).Create(ctx, h); err != nil {
			return err
		}
		return s.audits.WithTx(tx).Create(ctx, newAudit(user, model.ActionRequestApproval, model.EntityOutbound,
			h.ID, "membuat dan mengajukan "+h.ReferenceNumber))
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, user, h.ID)
}

func (s *OutboundService) Update(ctx context.Context, user *auth.User, id uuid.UUID, req dto.OutboundRequest) (*model.Outbound, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		h, err := s.lock(ctx, tx, user, id)
		if err != nil {
			return err
		}
		if h.Status == model.StatusApproved {
			return apperror.Conflict("out-bound yang sudah approved tidak bisa diubah (BR-009)")
		}
		if h.CreatedByUserID != user.ID {
			return apperror.Forbidden("hanya pembuat pengajuan yang bisa merevisi")
		}
		items, err := s.buildItems(ctx, h.WarehouseID, req)
		if err != nil {
			return err
		}
		repo := s.outbounds.WithTx(tx)
		if err := repo.ReplaceItems(ctx, h.ID, items); err != nil {
			return err
		}
		if err := repo.UpdateHeader(ctx, h.ID, map[string]any{
			"notes": req.Notes, "status": model.StatusPending, "submitted_at": time.Now(),
			"rejection_note": nil, "reviewed_at": nil, "reviewed_by_user_id": nil,
		}); err != nil {
			return err
		}
		return s.audits.WithTx(tx).Create(ctx, newAudit(user, model.ActionRequestApproval, model.EntityOutbound,
			h.ID, "merevisi dan mengajukan ulang "+h.ReferenceNumber))
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, user, id)
}

// Approve: stok di setiap rak berkurang (BR-004, BR-005).
func (s *OutboundService) Approve(ctx context.Context, user *auth.User, id uuid.UUID) (*model.Outbound, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		h, err := s.lock(ctx, tx, user, id)
		if err != nil {
			return err
		}
		if h.Status != model.StatusPending {
			return apperror.Conflict(fmt.Sprintf("hanya out-bound pending yang bisa di-approve (status: %s)", h.Status))
		}
		items, err := s.outbounds.WithTx(tx).ListItems(ctx, h.ID)
		if err != nil {
			return err
		}
		changes := make([]StockChange, 0, len(items))
		for _, it := range items {
			itemID := it.ID
			changes = append(changes, StockChange{
				WarehouseID: h.WarehouseID, LocationID: it.WarehouseLocationID, ProductID: it.ProductID,
				Direction: model.DirectionOut, Quantity: it.Quantity,
				MovementType: model.MovementOutbound, SourceID: h.ID, SourceItemID: &itemID,
				Label: fmt.Sprintf("%s di %s", it.Product.SKU, it.WarehouseLocation.Code),
			})
		}
		if err := s.stock.Apply(ctx, tx, user, changes); err != nil {
			return err
		}
		if err := s.outbounds.WithTx(tx).UpdateHeader(ctx, h.ID, map[string]any{
			"status": model.StatusApproved, "reviewed_at": time.Now(), "reviewed_by_user_id": user.ID,
		}); err != nil {
			return err
		}
		return s.audits.WithTx(tx).Create(ctx, newAudit(user, model.ActionApprove, model.EntityOutbound,
			h.ID, "menyetujui "+h.ReferenceNumber))
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, user, id)
}

func (s *OutboundService) Reject(ctx context.Context, user *auth.User, id uuid.UUID, note string) (*model.Outbound, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		h, err := s.lock(ctx, tx, user, id)
		if err != nil {
			return err
		}
		if h.Status != model.StatusPending {
			return apperror.Conflict(fmt.Sprintf("hanya out-bound pending yang bisa di-reject (status: %s)", h.Status))
		}
		if err := s.outbounds.WithTx(tx).UpdateHeader(ctx, h.ID, map[string]any{
			"status": model.StatusRejected, "rejection_note": note,
			"reviewed_at": time.Now(), "reviewed_by_user_id": user.ID,
		}); err != nil {
			return err
		}
		return s.audits.WithTx(tx).Create(ctx, newAudit(user, model.ActionReject, model.EntityOutbound, h.ID, note))
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, user, id)
}

func (s *OutboundService) lock(ctx context.Context, tx *gorm.DB, user *auth.User, id uuid.UUID) (*model.Outbound, error) {
	h, err := s.outbounds.WithTx(tx).FindByIDForUpdate(ctx, user.WarehouseScope(), id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.NotFound("out-bound")
	}
	return h, err
}

func (s *OutboundService) buildItems(ctx context.Context, warehouseID uuid.UUID, req dto.OutboundRequest) ([]model.OutboundItem, error) {
	var productIDs, locationIDs []uuid.UUID
	for _, it := range req.Items {
		productIDs = append(productIDs, it.ProductID)
		locationIDs = append(locationIDs, it.WarehouseLocationID)
	}
	master, err := loadMasterData(ctx, s.lookups, warehouseID, productIDs, locationIDs)
	if err != nil {
		return nil, err
	}

	errs := map[string]string{}
	seen := map[string]bool{}
	items := make([]model.OutboundItem, 0, len(req.Items))
	for i, it := range req.Items {
		field := fmt.Sprintf("items[%d]", i)
		master.checkProduct(errs, field+".product_id", it.ProductID)
		master.checkLocation(errs, field+".warehouse_location_id", it.WarehouseLocationID)
		key := it.WarehouseLocationID.String() + "|" + it.ProductID.String()
		if seen[key] {
			errs[field] = "produk yang sama di lokasi yang sama sudah ada di item lain"
		}
		seen[key] = true

		items = append(items, model.OutboundItem{
			ProductID: it.ProductID, WarehouseLocationID: it.WarehouseLocationID,
			Quantity: it.Quantity, Notes: it.Notes,
		})
	}
	if len(errs) > 0 {
		return nil, apperror.Validation(errs)
	}
	return items, nil
}
