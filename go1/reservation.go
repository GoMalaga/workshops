package reservations

import "errors"

// Status represents where a reservation is in its lifecycle.
type Status int

const (
	Held Status = iota
	Confirmed
	Canceled
)

// String makes a Status print with its name.
// NOTE: value receiver (s Status) because it only READS, it mutates nothing.
// Use it as a template for the methods you write below.
func (s Status) String() string {
	switch s {
	case Held:
		return "Held"
	case Confirmed:
		return "Confirmed"
	case Canceled:
		return "Canceled"
	default:
		return "Unknown"
	}
}

// ErrIllegalTransition is returned when an invalid state change is attempted.
var ErrIllegalTransition = errors.New("illegal state transition")

type Product struct {
	ID    string
	Name  string
	Stock int
}

type Reservation struct {
	ID        string
	ProductID string
	Quantity  int
	Status    Status
}

// confirm moves the reservation from Held to Confirmed.
// Only legal from Held; from any other state it returns ErrIllegalTransition.
// Hint: POINTER receiver (r *Reservation) because you MUTATE the state.
func (r *Reservation) confirm() error {
	if r.Status != Held {
		return ErrIllegalTransition
	}
	r.Status = Confirmed
	return nil
}

// cancel moves the reservation from Held to Canceled.
// Only legal from Held; from any other state it returns ErrIllegalTransition.
func (r *Reservation) cancel() error {
	if r.Status != Held {
		return ErrIllegalTransition
	}
	r.Status = Canceled
	return nil
}
