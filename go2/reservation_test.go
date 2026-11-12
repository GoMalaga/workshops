package reservations

import (
	"errors"
	"testing"
)

// newHeldReservation es el punto de partida de todos los tests de abajo: una
// reserva de 2 unidades, Held.
func newHeldReservation() *Reservation {
	return &Reservation{ID: "r1", ProductID: "p1", Quantity: 2, Status: Held}
}

// Un test con tabla: un bucle y una fila por caso. Es la forma habitual de
// escribir tests en Go.
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

// Una reserva Held se puede confirmar: sin error, y el estado cambia.
func TestConfirm_fromHeld(t *testing.T) {
	r := newHeldReservation()
	if err := r.confirm(); err != nil {
		t.Fatalf("expected confirm to succeed, got: %v", err)
	}
	if r.Status != Confirmed {
		t.Errorf("expected Confirmed, got: %v", r.Status)
	}
}

// Una reserva Held se puede cancelar: sin error, y el estado cambia.
func TestCancel_fromHeld(t *testing.T) {
	r := newHeldReservation()
	if err := r.cancel(); err != nil {
		t.Fatalf("expected cancel to succeed, got: %v", err)
	}
	if r.Status != Canceled {
		t.Errorf("expected Canceled, got: %v", r.Status)
	}
}

// Los cuatro tests de abajo cubren todas las transiciones que no se permiten.
// Cuando una reserva ya está Confirmed o Canceled, confirm y cancel devuelven
// ErrIllegalTransition.
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
