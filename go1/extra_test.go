package reservations

import (
	"errors"
	"testing"
)

// Tests for block 3. They are not in the scaffold on purpose: the student
// decides the rule for non-positive quantities and writes the test themselves.

func TestReserve_zeroQuantity_isRejected(t *testing.T) {
	s := newStoreWithStock(10)
	if _, err := s.Reserve("p1", 0); !errors.Is(err, ErrInvalidQuantity) {
		t.Errorf("expected ErrInvalidQuantity, got: %v", err)
	}
}

func TestReserve_negativeQuantity_isRejected(t *testing.T) {
	s := newStoreWithStock(10)
	_, err := s.Reserve("p1", -3)
	if !errors.Is(err, ErrInvalidQuantity) {
		t.Errorf("expected ErrInvalidQuantity, got: %v", err)
	}
	if got := s.products["p1"].Stock; got != 10 {
		t.Errorf("expected stock untouched at 10, got: %d", got)
	}
}

func TestStock_followsReserveAndCancel(t *testing.T) {
	s := newStoreWithStock(10)
	r, _ := s.Reserve("p1", 4)
	if got, _ := s.Stock("p1"); got != 6 {
		t.Errorf("expected stock 6 after reserving, got: %d", got)
	}
	_ = s.Cancel(r.ID)
	if got, _ := s.Stock("p1"); got != 10 {
		t.Errorf("expected stock 10 after canceling, got: %d", got)
	}
}

func TestStock_unknownProduct(t *testing.T) {
	s := NewStore()
	if _, err := s.Stock("ghost"); !errors.Is(err, ErrProductNotFound) {
		t.Errorf("expected ErrProductNotFound, got: %v", err)
	}
}
