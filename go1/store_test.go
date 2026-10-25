package reservations

import (
	"errors"
	"testing"
)

func newStoreWithStock(stock int) *Store {
	s := NewStore()
	s.AddProduct(&Product{ID: "p1", Name: "Ticket", Stock: stock})
	return s
}

func TestReserve_withinStock(t *testing.T) {
	s := newStoreWithStock(10)
	r, err := s.Reserve("p1", 3)
	if err != nil {
		t.Fatalf("expected reserve to succeed, got: %v", err)
	}
	if r.Status != Held {
		t.Errorf("expected Held, got: %v", r.Status)
	}
	if got := s.products["p1"].Stock; got != 7 {
		t.Errorf("expected stock 7, got: %d", got)
	}
}

func TestReserve_exactlyAllStock(t *testing.T) {
	s := newStoreWithStock(5)
	if _, err := s.Reserve("p1", 5); err != nil {
		t.Fatalf("expected reserve to succeed, got: %v", err)
	}
	if got := s.products["p1"].Stock; got != 0 {
		t.Errorf("expected stock 0, got: %d", got)
	}
}

func TestReserve_overStock_isRejected(t *testing.T) {
	s := newStoreWithStock(2)
	_, err := s.Reserve("p1", 3)
	if !errors.Is(err, ErrOutOfStock) {
		t.Errorf("expected ErrOutOfStock, got: %v", err)
	}
	if got := s.products["p1"].Stock; got != 2 {
		t.Errorf("expected stock untouched at 2, got: %d", got)
	}
}

func TestReserve_zeroStock_isRejected(t *testing.T) {
	s := newStoreWithStock(0)
	if _, err := s.Reserve("p1", 1); !errors.Is(err, ErrOutOfStock) {
		t.Errorf("expected ErrOutOfStock, got: %v", err)
	}
}

func TestReserve_unknownProduct(t *testing.T) {
	s := NewStore()
	if _, err := s.Reserve("ghost", 1); !errors.Is(err, ErrProductNotFound) {
		t.Errorf("expected ErrProductNotFound, got: %v", err)
	}
}

func TestStoreCancel_returnsStock(t *testing.T) {
	s := newStoreWithStock(10)
	r, err := s.Reserve("p1", 4) // stock now 6
	if err != nil {
		t.Fatalf("reserve failed: %v", err)
	}
	if err := s.Cancel(r.ID); err != nil {
		t.Fatalf("expected cancel to succeed, got: %v", err)
	}
	if got := s.products["p1"].Stock; got != 10 {
		t.Errorf("expected stock restored to 10, got: %d", got)
	}
}

func TestStoreConfirm_keepsStockDecremented(t *testing.T) {
	s := newStoreWithStock(10)
	r, err := s.Reserve("p1", 4) // stock now 6
	if err != nil {
		t.Fatalf("reserve failed: %v", err)
	}
	if err := s.Confirm(r.ID); err != nil {
		t.Fatalf("expected confirm to succeed, got: %v", err)
	}
	if got := s.products["p1"].Stock; got != 6 {
		t.Errorf("expected stock to stay at 6, got: %d", got)
	}
}

func TestStoreCancel_twice_keepsStock(t *testing.T) {
	s := newStoreWithStock(10)
	r, err := s.Reserve("p1", 4) // stock now 6
	if err != nil {
		t.Fatalf("reserve failed: %v", err)
	}
	if err := s.Cancel(r.ID); err != nil {
		t.Fatalf("expected the first cancel to succeed, got: %v", err)
	}
	if err := s.Cancel(r.ID); !errors.Is(err, ErrIllegalTransition) {
		t.Errorf("expected ErrIllegalTransition on the second cancel, got: %v", err)
	}
	if got := s.products["p1"].Stock; got != 10 {
		t.Errorf("expected stock to stay at 10, got: %d", got)
	}
}

func TestStoreCancel_unknownReservation(t *testing.T) {
	s := newStoreWithStock(10)
	if err := s.Cancel("ghost"); !errors.Is(err, ErrReservationNotFound) {
		t.Errorf("expected ErrReservationNotFound, got: %v", err)
	}
}

func TestReservations_byProduct_sortedByID(t *testing.T) {
	s := NewStore()
	s.AddProduct(&Product{ID: "p1", Name: "Ticket", Stock: 10})
	s.AddProduct(&Product{ID: "p2", Name: "Poster", Stock: 10})
	s.Reserve("p1", 1) // r1
	s.Reserve("p2", 1) // r2
	s.Reserve("p1", 2) // r3
	s.Reserve("p1", 3) // r4
	got := s.Reservations("p1")
	if len(got) != 3 {
		t.Fatalf("expected 3 reservations for p1, got: %d", len(got))
	}
	for i, want := range []string{"r1", "r3", "r4"} {
		if got[i].ID != want {
			t.Errorf("position %d: expected %s, got: %s", i, want, got[i].ID)
		}
	}
}

// An empty slice is not a nil slice: len() is 0 for both, but in session 2
// nil becomes null in JSON and an empty slice becomes [].
func TestReservations_unknownProduct_isEmpty(t *testing.T) {
	s := newStoreWithStock(10)
	if got := s.Reservations("ghost"); got == nil || len(got) != 0 {
		t.Errorf("expected an empty slice, got: %#v", got)
	}
}
