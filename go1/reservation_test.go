package reservations

import (
	"errors"
	"testing"
)

func newHeldReservation() *Reservation {
	return &Reservation{ID: "r1", ProductID: "p1", Quantity: 2, Status: Held}
}

// A table-driven test: one loop, one row per case. The Go way to test.
func TestStatus_String(t *testing.T) {
	cases := []struct {
		status Status
		want   string
	}{
		{Held, "Held"},
		{Confirmed, "Confirmed"},
		{Canceled, "Canceled"},
		{Status(99), "Unknown"},
	}
	for _, c := range cases {
		if got := c.status.String(); got != c.want {
			t.Errorf("Status(%d).String() = %q, want %q", c.status, got, c.want)
		}
	}
}

func TestConfirm_fromHeld(t *testing.T) {
	r := newHeldReservation()
	if err := r.confirm(); err != nil {
		t.Fatalf("expected confirm to succeed, got: %v", err)
	}
	if r.Status != Confirmed {
		t.Errorf("expected Confirmed, got: %v", r.Status)
	}
}

func TestCancel_fromHeld(t *testing.T) {
	r := newHeldReservation()
	if err := r.cancel(); err != nil {
		t.Fatalf("expected cancel to succeed, got: %v", err)
	}
	if r.Status != Canceled {
		t.Errorf("expected Canceled, got: %v", r.Status)
	}
}

func TestConfirm_alreadyConfirmed_isIllegal(t *testing.T) {
	r := newHeldReservation()
	_ = r.confirm()
	if err := r.confirm(); !errors.Is(err, ErrIllegalTransition) {
		t.Errorf("expected ErrIllegalTransition, got: %v", err)
	}
}

func TestCancel_alreadyConfirmed_isIllegal(t *testing.T) {
	r := newHeldReservation()
	_ = r.confirm()
	if err := r.cancel(); !errors.Is(err, ErrIllegalTransition) {
		t.Errorf("expected ErrIllegalTransition, got: %v", err)
	}
}

func TestConfirm_alreadyCanceled_isIllegal(t *testing.T) {
	r := newHeldReservation()
	_ = r.cancel()
	if err := r.confirm(); !errors.Is(err, ErrIllegalTransition) {
		t.Errorf("expected ErrIllegalTransition, got: %v", err)
	}
}

func TestCancel_alreadyCanceled_isIllegal(t *testing.T) {
	r := newHeldReservation()
	_ = r.cancel()
	if err := r.cancel(); !errors.Is(err, ErrIllegalTransition) {
		t.Errorf("expected ErrIllegalTransition, got: %v", err)
	}
}
