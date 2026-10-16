package service

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"warehouse-management/internal/auth"
	"warehouse-management/internal/model"
	"warehouse-management/internal/repository"
	"warehouse-management/pkg/apperror"
	"warehouse-management/pkg/pagination"
)

// StockChange adalah satu perubahan stok yang diminta oleh sebuah transaksi.
type StockChange struct {
	WarehouseID uuid.UUID
	LocationID  uuid.UUID
	ProductID   uuid.UUID
	Direction   string // model.DirectionIn / model.DirectionOut
	Quantity    int    // selalu positif; arah ditentukan Direction

	MovementType string     // model.MovementStockAdjustment, dst.
	SourceID     uuid.UUID  // id header transaksi
	SourceItemID *uuid.UUID // id item transaksi

	Label string // untuk pesan error, mis. "LP-001 di RAK-A-01"
}

// StockService adalah SATU-SATUNYA tempat yang boleh mengubah saldo stok.
type StockService struct {
	stocks *repository.StockRepository
}

func NewStockService(stocks *repository.StockRepository) *StockService {
	return &StockService{stocks: stocks}
}

// Apply menjalankan beberapa perubahan stok di dalam transaksi tx milik pemanggil.
// Kalau satu perubahan gagal, pemanggil mengembalikan error dan SEMUA perubahan
// di transaksi itu ikut rollback (NFR-007).
func (s *StockService) Apply(ctx context.Context, tx *gorm.DB, user *auth.User, changes []StockChange) error {
	// Kunci baris selalu dalam urutan yang sama (lokasi, lalu produk).
	// Tanpa ini, dua transaksi yang mengunci A-B dan B-A bisa saling menunggu (deadlock).
	sorted := make([]StockChange, len(changes))
	copy(sorted, changes)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].LocationID != sorted[j].LocationID {
			return sorted[i].LocationID.String() < sorted[j].LocationID.String()
		}
		return sorted[i].ProductID.String() < sorted[j].ProductID.String()
	})

	stocks := s.stocks.WithTx(tx)
	now := time.Now()

	for _, ch := range sorted {
		if ch.Quantity <= 0 {
			return fmt.Errorf("quantity perubahan stok harus positif, dapat %d", ch.Quantity)
		}

		// 1. Kunci saldo. Transaksi lain pada saldo yang sama menunggu di baris ini.
		balance, err := stocks.LockBalance(ctx, ch.WarehouseID, ch.LocationID, ch.ProductID)
		if err != nil {
			return err
		}

		// 2. Hitung saldo baru dan pastikan tidak negatif (BR-005).
		newQty := balance.Quantity + ch.Quantity
		if ch.Direction == model.DirectionOut {
			newQty = balance.Quantity - ch.Quantity
		}
		if newQty < 0 {
			return apperror.New(http.StatusUnprocessableEntity, fmt.Sprintf(
				"stok tidak cukup untuk %s: tersedia %d, dibutuhkan %d",
				ch.Label, balance.Quantity, ch.Quantity))
		}

		// 3. Simpan saldo baru.
		balance.Quantity = newQty
		if err := stocks.UpdateQuantity(ctx, balance); err != nil {
			return err
		}

		// 4. Tulis ledger (BR-006).
		if err := stocks.CreateMovement(ctx, &model.StockMovement{
			WarehouseID:         ch.WarehouseID,
			WarehouseLocationID: ch.LocationID,
			ProductID:           ch.ProductID,
			MovementType:        ch.MovementType,
			MovementDirection:   ch.Direction,
			Quantity:            ch.Quantity,
			StockAfter:          newQty,
			SourceEntityType:    ch.MovementType,
			SourceEntityID:      ch.SourceID,
			SourceItemID:        ch.SourceItemID,
			OccurredAt:          now,
			CreatedByUserID:     user.ID,
		}); err != nil {
			return err
		}
	}
	return nil
}

// ListBalances menampilkan saldo stok, dibatasi gudang milik user.
func (s *StockService) ListBalances(ctx context.Context, user *auth.User, f repository.StockFilter, p pagination.Params) ([]repository.StockBalanceRow, int64, error) {
	return s.stocks.ListBalances(ctx, user.WarehouseScope(), f, p)
}
