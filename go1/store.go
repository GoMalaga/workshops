package reservations

import (
	"errors"
	"fmt"
	"sort"
)

var (
	ErrProductNotFound     = errors.New("product not found")
	ErrOutOfStock          = errors.New("out of stock")
	ErrReservationNotFound = errors.New("reservation not found")
	ErrInvalidQuantity     = errors.New("invalid quantity")
)

// Store holds products and reservations in memory.
type Store struct {
	products     map[string]*Product
	reservations map[string]*Reservation
	seq          int
}

func NewStore() *Store {
	return &Store{
		products:     make(map[string]*Product),
		reservations: make(map[string]*Reservation),
	}
}

// AddProduct seeds a product. Given, don't touch it.
func (s *Store) AddProduct(p *Product) {
	s.products[p.ID] = p
}

// Stock reports how many units of a product are still available to reserve.
// Block 3: not in the scaffold. The taquilla has to read stock from outside the
// package and products is unexported, hence an exported method.
// ErrProductNotFound if the product doesn't exist.
func (s *Store) Stock(productID string) (int, error) {
	p := s.products[productID]
	if p == nil {
		return 0, ErrProductNotFound
	}
	return p.Stock, nil
}

// nextID returns a fresh reservation ID: r1, r2, r3...
func (s *Store) nextID() string {
	s.seq++
	return fmt.Sprintf("r%d", s.seq)
}

// Confirm looks up the reservation and delegates the transition to it.
// WORKED EXAMPLE: notice the shape: find, guard nil, delegate.
// Your Reserve and Cancel follow the same pattern.
func (s *Store) Confirm(reservationID string) error {
	r := s.reservations[reservationID]
	if r == nil {
		return ErrReservationNotFound
	}
	return r.confirm()
}

// Reserve checks stock, decrements it, and records a Held reservation.
// INVARIANT: never let stock go below zero.
// ErrProductNotFound if the product doesn't exist; ErrOutOfStock if not enough stock.
// Block 3: ErrInvalidQuantity if qty is zero or negative. Not in the scaffold:
// without this guard Reserve("p1", -3) ADDS stock and every test stays green.
func (s *Store) Reserve(productID string, qty int) (*Reservation, error) {
	if qty <= 0 {
		return nil, ErrInvalidQuantity
	}
	p := s.products[productID]
	if p == nil {
		return nil, ErrProductNotFound
	}
	if p.Stock < qty {
		return nil, ErrOutOfStock
	}
	p.Stock -= qty
	r := &Reservation{ID: s.nextID(), ProductID: productID, Quantity: qty, Status: Held}
	s.reservations[r.ID] = r
	return r, nil
}

// Cancel releases a Held reservation and RETURNS its quantity to the product's stock.
// Only return stock if the cancel transition was legal.
// ErrReservationNotFound if the reservation doesn't exist.
func (s *Store) Cancel(reservationID string) error {
	r := s.reservations[reservationID]
	if r == nil {
		return ErrReservationNotFound
	}
	if err := r.cancel(); err != nil {
		return err
	}
	s.products[r.ProductID].Stock += r.Quantity
	return nil
}

// Reservations returns every reservation of a product, sorted by ID.
// A map has no order: range over it, collect into a slice, then sort.
// Unknown product or no reservations: an empty slice (not nil), never an error.
func (s *Store) Reservations(productID string) []*Reservation {
	rs := make([]*Reservation, 0)
	for _, r := range s.reservations {
		if r.ProductID == productID {
			rs = append(rs, r)
		}
	}
	sort.Slice(rs, func(i, j int) bool { return rs[i].ID < rs[j].ID })
	return rs
}
