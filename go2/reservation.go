package reservations

import (
	"errors"
	"fmt"
)

// Status dice en qué punto de su ciclo de vida está una reserva.
type Status int

const (
	Held Status = iota
	Confirmed
	Canceled
)

// String hace que un Status se imprima con su nombre.
// NOTA: receptor por valor (s Status) porque solo LEE, no modifica nada.
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

// Los errores del dominio. El Store los devuelve tal cual y la API convierte
// cada uno en un código de estado.
var (
	ErrProductNotFound     = errors.New("product not found")
	ErrOutOfStock          = errors.New("out of stock")
	ErrReservationNotFound = errors.New("reservation not found")
	ErrInvalidQuantity     = errors.New("invalid quantity")
	ErrIllegalTransition   = errors.New("illegal state transition")
)

// MarshalText hace que un Status viaje por JSON con su nombre ("Held"), no
// como un número. encoding/json lo llama en cualquier tipo que lo tenga.
func (s Status) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText es el camino de vuelta: "Held" en un cuerpo JSON pasa a ser
// Held. Receptor de puntero: tiene que escribir en el Status.
func (s *Status) UnmarshalText(text []byte) error {
	switch string(text) {
	case "Held":
		*s = Held
	case "Confirmed":
		*s = Confirmed
	case "Canceled":
		*s = Canceled
	default:
		return fmt.Errorf("unknown status %q", text)
	}
	return nil
}

// Las etiquetas json dan a los campos el nombre con el que los muestra la API:
// {"id": "p1", "stock": 10}.
type Product struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Stock int    `json:"stock"`
}

type Reservation struct {
	ID        string `json:"id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Status    Status `json:"status"`
}

// confirm pasa la reserva de Held a Confirmed.
// Solo es legal desde Held; desde cualquier otro estado devuelve
// ErrIllegalTransition.
// Receptor de PUNTERO (r *Reservation) porque MODIFICA el estado.
func (r *Reservation) confirm() error {
	if r.Status != Held {
		return ErrIllegalTransition
	}
	r.Status = Confirmed
	return nil
}

// cancel pasa la reserva de Held a Canceled.
// Solo es legal desde Held; desde cualquier otro estado devuelve
// ErrIllegalTransition.
func (r *Reservation) cancel() error {
	if r.Status != Held {
		return ErrIllegalTransition
	}
	r.Status = Canceled
	return nil
}
