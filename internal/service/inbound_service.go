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

// InboundService. Berbeda dari transaksi lain:
//   - Approve TIDAK mengubah stok (barang baru diterima, belum ditaruh di rak).
//   - Putaway: staff menentukan rak untuk setiap item -> stok bertambah, status completed.
type InboundService struct {
	db         *gorm.DB
	inbounds   *repository.InboundRepository
	warehouses *repository.WarehouseRepository
	lookups    *repository.LookupRepository
	sequences  *repository.SequenceRepository
	audits     *repository.AuditRepository
	stock      *StockService
}

func NewInboundService(db *gorm.DB, inbounds *repository.InboundRepository, warehouses *repository.WarehouseRepository,
	lookups *repository.LookupRepository, sequences *repository.SequenceRepository, audits *repository.AuditRepository, stock *StockService) *InboundService {
	return &InboundService{db: db, inbounds: inbounds, warehouses: warehouses, lookups: lookups, sequences: sequences, audits: audits, stock: stock}
}

func (s *InboundService) List(ctx context.Context, user *auth.User, f repository.TransactionFilter, p pagination.Params) ([]model.Inbound, int64, error) {
	return s.inbounds.List(ctx, user.WarehouseScope(), f, p)
}

func (s *InboundService) Get(ctx context.Context, user *auth.User, id uuid.UUID) (*model.Inbound, error) {
	h, err := s.inbounds.FindByID(ctx, user.WarehouseScope(), id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.NotFound("in-bound")
	}
	return h, err
}

func (s *InboundService) Create(ctx context.Context, user *auth.User, req dto.InboundRequest) (*model.Inbound, error) {
	if err := checkWarehouseAccess(ctx, s.warehouses, user, req.WarehouseID); err != nil {
		return nil, err
	}
	items, err := s.buildItems(ctx, req)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	h := &model.Inbound{
		WarehouseID:         req.WarehouseID,
		PurchaseOrderNumber: req.PurchaseOrderNumber,
		Status:              model.StatusPending,
		Notes:               req.Notes,
		SubmittedAt:         &now,
		CreatedByUserID:     user.ID,
		Items:               items,
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ref, err := s.sequences.WithTx(tx).Next(ctx, "INB", now)
		if err != nil {
			return err
		}
		h.ReferenceNumber = ref
		if err := s.inbounds.WithTx(tx).Create(ctx, h); err != nil {
			return err
		}
		return s.audits.WithTx(tx).Create(ctx, newAudit(user, model.ActionRequestApproval, model.EntityInbound,
			h.ID, "membuat dan mengajukan "+h.ReferenceNumber))
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, user, h.ID)
}

func (s *InboundService) Update(ctx context.Context, user *auth.User, id uuid.UUID, req dto.InboundRequest) (*model.Inbound, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		h, err := s.lock(ctx, tx, user, id)
		if err != nil {
			return err
		}
		if h.Status != model.StatusPending && h.Status != model.StatusRejected {
			return apperror.Conflict(fmt.Sprintf("in-bound berstatus %s tidak bisa diubah (BR-009)", h.Status))
		}
		if h.CreatedByUserID != user.ID {
			return apperror.Forbidden("hanya pembuat pengajuan yang bisa merevisi")
		}
		items, err := s.buildItems(ctx, req)
		if err != nil {
			return err
		}
		repo := s.inbounds.WithTx(tx)
		if err := repo.ReplaceItems(ctx, h.ID, items); err != nil {
			return err
		}
		if err := repo.UpdateHeader(ctx, h.ID, map[string]any{
			"purchase_order_number": req.PurchaseOrderNumber, "notes": req.Notes,
			"status": model.StatusPending, "submitted_at": time.Now(),
			"rejection_note": nil, "reviewed_at": nil, "reviewed_by_user_id": nil,
		}); err != nil {
			return err
		}
		return s.audits.WithTx(tx).Create(ctx, newAudit(user, model.ActionRequestApproval, model.EntityInbound,
			h.ID, "merevisi dan mengajukan ulang "+h.ReferenceNumber))
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, user, id)
}

// Approve: hanya mengubah status. Stok BELUM berubah sampai putaway.
func (s *InboundService) Approve(ctx context.Context, user *auth.User, id uuid.UUID) (*model.Inbound, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		h, err := s.lock(ctx, tx, user, id)
		if err != nil {
			return err
		}
		if h.Status != model.StatusPending {
			return apperror.Conflict(fmt.Sprintf("hanya in-bound pending yang bisa di-approve (status: %s)", h.Status))
		}
		if err := s.inbounds.WithTx(tx).UpdateHeader(ctx, h.ID, map[string]any{
			"status": model.StatusApproved, "reviewed_at": time.Now(), "reviewed_by_user_id": user.ID,
		}); err != nil {
			return err
		}
		return s.audits.WithTx(tx).Create(ctx, newAudit(user, model.ActionApprove, model.EntityInbound,
			h.ID, "menyetujui "+h.ReferenceNumber))
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, user, id)
}

func (s *InboundService) Reject(ctx context.Context, user *auth.User, id uuid.UUID, note string) (*model.Inbound, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		h, err := s.lock(ctx, tx, user, id)
		if err != nil {
			return err
		}
		if h.Status != model.StatusPending {
			return apperror.Conflict(fmt.Sprintf("hanya in-bound pending yang bisa di-reject (status: %s)", h.Status))
		}
		if err := s.inbounds.WithTx(tx).UpdateHeader(ctx, h.ID, map[string]any{
			"status": model.StatusRejected, "rejection_note": note,
			"reviewed_at": time.Now(), "reviewed_by_user_id": user.ID,
		}); err != nil {
			return err
		}
		return s.audits.WithTx(tx).Create(ctx, newAudit(user, model.ActionReject, model.EntityInbound, h.ID, note))
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, user, id)
}

// Putaway: staff menaruh barang In-Bound yang sudah approved ke rak.
// Setiap item WAJIB diberi rak; stok bertambah dan stock movement "inbound" tercatat.
func (s *InboundService) Putaway(ctx context.Context, user *auth.User, id uuid.UUID, req dto.PutawayRequest) (*model.Inbound, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Kunci header: putaway dua kali bersamaan -> yang kedua melihat status completed.
		h, err := s.lock(ctx, tx, user, id)
		if err != nil {
			return err
		}
		if h.Status != model.StatusApproved {
			return apperror.Conflict(fmt.Sprintf("putaway hanya untuk in-bound approved (status: %s)", h.Status))
		}

		repo := s.inbounds.WithTx(tx)
		items, err := repo.ListItems(ctx, h.ID)
		if err != nil {
			return err
		}

		// Petakan rak yang dikirim per item, lalu validasi.
		requested := map[uuid.UUID]uuid.UUID{} // inbound_item_id -> warehouse_location_id
		var locationIDs []uuid.UUID
		errs := map[string]string{}
		for i, it := range req.Items {
			if _, dup := requested[it.InboundItemID]; dup {
				errs[fmt.Sprintf("items[%d].inbound_item_id", i)] = "item ini dikirim lebih dari sekali"
			}
			requested[it.InboundItemID] = it.WarehouseLocationID
			locationIDs = append(locationIDs, it.WarehouseLocationID)
		}
		master, err := loadMasterData(ctx, s.lookups, h.WarehouseID, nil, locationIDs)
		if err != nil {
			return err
		}
		for i, it := range req.Items {
			master.checkLocation(errs, fmt.Sprintf("items[%d].warehouse_location_id", i), it.WarehouseLocationID)
		}
		itemIDs := map[uuid.UUID]bool{}
		for _, it := range items {
			itemIDs[it.ID] = true
			if _, ok := requested[it.ID]; !ok {
				errs["items"] = fmt.Sprintf("semua item harus diberi rak (%s belum)", it.Product.SKU)
			}
		}
		for i, it := range req.Items {
			if !itemIDs[it.InboundItemID] {
				errs[fmt.Sprintf("items[%d].inbound_item_id", i)] = "bukan item dari in-bound ini"
			}
		}
		if len(errs) > 0 {
			return apperror.Validation(errs)
		}

		// Simpan rak per item + susun perubahan stok (semua arah masuk).
		changes := make([]StockChange, 0, len(items))
		for _, it := range items {
			locationID := requested[it.ID]
			if err := repo.SetItemLocation(ctx, it.ID, locationID); err != nil {
				return err
			}
			itemID := it.ID
			changes = append(changes, StockChange{
				WarehouseID: h.WarehouseID, LocationID: locationID, ProductID: it.ProductID,
				Direction: model.DirectionIn, Quantity: it.Quantity,
				MovementType: model.MovementInbound, SourceID: h.ID, SourceItemID: &itemID,
				Label: fmt.Sprintf("%s di %s", it.Product.SKU, master.locations[locationID]),
			})
		}
		if err := s.stock.Apply(ctx, tx, user, changes); err != nil {
			return err
		}
		if err := repo.UpdateHeader(ctx, h.ID, map[string]any{
			"status": model.StatusCompleted, "putaway_at": time.Now(), "putaway_by_user_id": user.ID,
		}); err != nil {
			return err
		}
		return s.audits.WithTx(tx).Create(ctx, newAudit(user, model.ActionPutaway, model.EntityInbound,
			h.ID, "putaway "+h.ReferenceNumber))
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, user, id)
}

func (s *InboundService) lock(ctx context.Context, tx *gorm.DB, user *auth.User, id uuid.UUID) (*model.Inbound, error) {
	h, err := s.inbounds.WithTx(tx).FindByIDForUpdate(ctx, user.WarehouseScope(), id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.NotFound("in-bound")
	}
	return h, err
}

// buildItems: In-Bound belum punya rak saat dibuat, jadi hanya produk yang dicek.
func (s *InboundService) buildItems(ctx context.Context, req dto.InboundRequest) ([]model.InboundItem, error) {
	var productIDs []uuid.UUID
	for _, it := range req.Items {
		productIDs = append(productIDs, it.ProductID)
	}
	master, err := loadMasterData(ctx, s.lookups, uuid.Nil, productIDs, nil)
	if err != nil {
		return nil, err
	}

	errs := map[string]string{}
	seen := map[uuid.UUID]bool{}
	items := make([]model.InboundItem, 0, len(req.Items))
	for i, it := range req.Items {
		field := fmt.Sprintf("items[%d]", i)
		master.checkProduct(errs, field+".product_id", it.ProductID)
		if seen[it.ProductID] {
			errs[field] = "produk yang sama sudah ada di item lain"
		}
		seen[it.ProductID] = true
		items = append(items, model.InboundItem{ProductID: it.ProductID, Quantity: it.Quantity, Notes: it.Notes})
	}
	if len(errs) > 0 {
		return nil, apperror.Validation(errs)
	}
	return items, nil
}
