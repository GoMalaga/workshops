package reservations

import (
	"errors"
	"path/filepath"
	"testing"
)

// Tests del bloque 1. Cada test tiene su propio fichero de base de datos, en
// una carpeta temporal que Go borra al acabar el test: ningún test ve los
// datos de otro.

// newStore abre un Store sobre un fichero de base de datos nuevo y lo devuelve
// junto con la ruta de ese fichero. El Store se cierra al acabar el test.
func newStore(t *testing.T) (*Store, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "taquilla.db")
	s, err := OpenStore(path)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s, path
}

// newStoreWithStock es newStore más un producto, p1 "Ticket", con el stock que
// se le pasa. Casi todos los tests parten de aquí.
func newStoreWithStock(t *testing.T, stock int) *Store {
	t.Helper()
	s, _ := newStore(t)
	if err := s.AddProduct(&Product{ID: "p1", Name: "Ticket", Stock: stock}); err != nil {
		t.Fatalf("add product: %v", err)
	}
	return s
}

// mustReserve reserva y para el test si falla. Para los tests en los que la
// reserva es solo la preparación, no lo que se comprueba.
func mustReserve(t *testing.T, s *Store, productID string, qty int) *Reservation {
	t.Helper()
	r, err := s.Reserve(productID, qty)
	if err != nil {
		t.Fatalf("reserve failed: %v", err)
	}
	return r
}

// stockOf lee el stock de un producto y para el test si no puede.
func stockOf(t *testing.T, s *Store, productID string) int {
	t.Helper()
	n, err := s.Stock(productID)
	if err != nil {
		t.Fatalf("stock failed: %v", err)
	}
	return n
}

// Paso 1. Un producto añadido con stock 10 se puede leer después: Stock dice
// 10. Comprueba el INSERT de AddProduct con el SELECT de Stock, que viene dado.
func TestAddProduct_thenStock(t *testing.T) {
	s := newStoreWithStock(t, 10)
	if got := stockOf(t, s, "p1"); got != 10 {
		t.Errorf("expected stock 10 right after adding, got: %d", got)
	}
}

// Pedir el stock de un producto que no existe devuelve ErrProductNotFound, no
// un cero ni el error en crudo de la base de datos.
func TestStock_unknownProduct(t *testing.T) {
	s, _ := newStore(t)
	if _, err := s.Stock("ghost"); !errors.Is(err, ErrProductNotFound) {
		t.Errorf("expected ErrProductNotFound, got: %v", err)
	}
}

// Paso 2. Product devuelve los tres campos tal como se añadieron. Si hay dos
// columnas cambiadas de orden en el SELECT o en el Scan, este es el test que
// falla.
func TestProduct_afterAdd(t *testing.T) {
	s := newStoreWithStock(t, 10)
	p, err := s.Product("p1")
	if err != nil {
		t.Fatalf("product failed: %v", err)
	}
	if p.ID != "p1" || p.Name != "Ticket" || p.Stock != 10 {
		t.Errorf("expected p1 Ticket x10, got: %+v", *p)
	}
}

// Product con un ID que no existe devuelve ErrProductNotFound: el "no hay
// filas" de la base de datos tiene que convertirse en el error del dominio.
func TestProduct_unknown(t *testing.T) {
	s, _ := newStore(t)
	if _, err := s.Product("ghost"); !errors.Is(err, ErrProductNotFound) {
		t.Errorf("expected ErrProductNotFound, got: %v", err)
	}
}

// Añade primero p2 y después p1, y espera que Products devuelva p1 y luego
// p2: el orden lo da la consulta (ORDER BY), no el orden de inserción.
func TestProducts_sortedByID(t *testing.T) {
	s, _ := newStore(t)
	for _, p := range []*Product{{ID: "p2", Name: "Poster", Stock: 3}, {ID: "p1", Name: "Ticket", Stock: 10}} {
		if err := s.AddProduct(p); err != nil {
			t.Fatalf("add product: %v", err)
		}
	}
	got, err := s.Products()
	if err != nil {
		t.Fatalf("products failed: %v", err)
	}
	if len(got) != 2 || got[0].ID != "p1" || got[1].ID != "p2" {
		t.Errorf("expected [p1 p2], got: %v", productIDs(got))
	}
}

// Paso 3, el caso normal. Reservar 3 de 10 devuelve una reserva Held con el ID
// que eligió la base de datos, y deja el stock en 7.
func TestReserve_decrementsStock(t *testing.T) {
	s := newStoreWithStock(t, 10)
	r := mustReserve(t, s, "p1", 3)
	if r.Status != Held {
		t.Errorf("expected a Held reservation, got: %v", r.Status)
	}
	if r.ID == "" {
		t.Error("expected the reservation to carry the ID the database gave it")
	}
	if got := stockOf(t, s, "p1"); got != 7 {
		t.Errorf("expected stock 7 after reserving 3, got: %d", got)
	}
}

// Reservar 3 cuando solo hay 2 devuelve ErrOutOfStock y no escribe nada: el
// stock sigue en 2. La comprobación tiene que ir antes del UPDATE.
func TestReserve_overStock_isRejected(t *testing.T) {
	s := newStoreWithStock(t, 2)
	if _, err := s.Reserve("p1", 3); !errors.Is(err, ErrOutOfStock) {
		t.Errorf("expected ErrOutOfStock, got: %v", err)
	}
	if got := stockOf(t, s, "p1"); got != 2 {
		t.Errorf("expected stock untouched at 2, got: %d", got)
	}
}

// Reservar un producto que no existe devuelve ErrProductNotFound.
func TestReserve_unknownProduct(t *testing.T) {
	s := newStoreWithStock(t, 10)
	if _, err := s.Reserve("ghost", 1); !errors.Is(err, ErrProductNotFound) {
		t.Errorf("expected ErrProductNotFound, got: %v", err)
	}
}

// Una cantidad negativa devuelve ErrInvalidQuantity y no toca el stock. Sin
// esta comprobación, reservar -3 SUMARÍA 3 unidades al stock.
func TestReserve_invalidQuantity(t *testing.T) {
	s := newStoreWithStock(t, 10)
	if _, err := s.Reserve("p1", -3); !errors.Is(err, ErrInvalidQuantity) {
		t.Errorf("expected ErrInvalidQuantity, got: %v", err)
	}
	if got := stockOf(t, s, "p1"); got != 10 {
		t.Errorf("expected stock untouched at 10, got: %d", got)
	}
}

// Paso 4. Reservation con un ID que no existe devuelve ErrReservationNotFound.
func TestReservation_unknown(t *testing.T) {
	s, _ := newStore(t)
	if _, err := s.Reservation("42"); !errors.Is(err, ErrReservationNotFound) {
		t.Errorf("expected ErrReservationNotFound, got: %v", err)
	}
}

// Confirmar una reserva Held guarda el estado nuevo: leída otra vez de la base
// de datos, dice Confirmed. El stock sigue en 6: confirmar no lo mueve.
func TestConfirm_keepsStockDecremented(t *testing.T) {
	s := newStoreWithStock(t, 10)
	r := mustReserve(t, s, "p1", 4)
	if err := s.Confirm(r.ID); err != nil {
		t.Fatalf("confirm failed: %v", err)
	}
	got, err := s.Reservation(r.ID)
	if err != nil {
		t.Fatalf("reservation failed: %v", err)
	}
	if got.Status != Confirmed {
		t.Errorf("expected Confirmed, got: %v", got.Status)
	}
	if stock := stockOf(t, s, "p1"); stock != 6 {
		t.Errorf("expected stock 6 after confirming, got: %d", stock)
	}
}

// Paso 5. Cancelar una reserva Held guarda el estado Canceled y devuelve sus
// 4 unidades: el stock vuelve a 10.
func TestCancel_returnsStock(t *testing.T) {
	s := newStoreWithStock(t, 10)
	r := mustReserve(t, s, "p1", 4)
	if err := s.Cancel(r.ID); err != nil {
		t.Fatalf("cancel failed: %v", err)
	}
	got, _ := s.Reservation(r.ID)
	if got.Status != Canceled {
		t.Errorf("expected Canceled, got: %v", got.Status)
	}
	if stock := stockOf(t, s, "p1"); stock != 10 {
		t.Errorf("expected stock back at 10 after canceling, got: %d", stock)
	}
}

// La segunda cancelación de la misma reserva devuelve ErrIllegalTransition y
// no puede devolver el stock otra vez: se queda en 10, no en 14.
func TestCancel_twice_keepsStock(t *testing.T) {
	s := newStoreWithStock(t, 10)
	r := mustReserve(t, s, "p1", 4)
	_ = s.Cancel(r.ID)
	if err := s.Cancel(r.ID); !errors.Is(err, ErrIllegalTransition) {
		t.Errorf("expected ErrIllegalTransition on the second cancel, got: %v", err)
	}
	if stock := stockOf(t, s, "p1"); stock != 10 {
		t.Errorf("expected stock 10, not returned twice, got: %d", stock)
	}
}

// Cancelar una reserva que no existe devuelve ErrReservationNotFound.
func TestCancel_unknownReservation(t *testing.T) {
	s := newStoreWithStock(t, 10)
	if err := s.Cancel("999"); !errors.Is(err, ErrReservationNotFound) {
		t.Errorf("expected ErrReservationNotFound, got: %v", err)
	}
}

// Paso 6. Tres reservas: dos de p1 y, entre medias, una de p2. Reservations de
// p1 devuelve solo las dos suyas, la más antigua primero: comprueba el WHERE y
// el ORDER BY.
func TestReservations_byProduct_sortedByID(t *testing.T) {
	s := newStoreWithStock(t, 10)
	if err := s.AddProduct(&Product{ID: "p2", Name: "Poster", Stock: 5}); err != nil {
		t.Fatalf("add product: %v", err)
	}
	r1 := mustReserve(t, s, "p1", 1)
	mustReserve(t, s, "p2", 1)
	r3 := mustReserve(t, s, "p1", 2)
	got, err := s.Reservations("p1")
	if err != nil {
		t.Fatalf("reservations failed: %v", err)
	}
	if len(got) != 2 || got[0].ID != r1.ID || got[1].ID != r3.ID {
		t.Errorf("expected [%s %s], got: %v", r1.ID, r3.ID, ids(got))
	}
}

// Un producto sin reservas da un slice vacío y ningún error. El slice no puede
// ser nil: nil viajaría como null en el JSON del bloque 2, no como [].
func TestReservations_unknownProduct_isEmpty(t *testing.T) {
	s := newStoreWithStock(t, 10)
	got, err := s.Reservations("ghost")
	if err != nil {
		t.Fatalf("reservations failed: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("expected an empty slice, got: %v", got)
	}
}

// Cierre del bloque, su razón de ser: se cierra el Store, se abre otra vez el
// mismo fichero, y el stock y la reserva siguen ahí.
func TestStore_survivesReopen(t *testing.T) {
	s, path := newStore(t)
	if err := s.AddProduct(&Product{ID: "p1", Name: "Ticket", Stock: 10}); err != nil {
		t.Fatalf("add product: %v", err)
	}
	r := mustReserve(t, s, "p1", 3)
	if err := s.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	reopened, err := OpenStore(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer reopened.Close()
	if got := stockOf(t, reopened, "p1"); got != 7 {
		t.Errorf("expected stock 7 after reopening, got: %d", got)
	}
	again, err := reopened.Reservation(r.ID)
	if err != nil {
		t.Fatalf("expected reservation %s to survive the restart, got: %v", r.ID, err)
	}
	if again.Status != Held || again.Quantity != 3 {
		t.Errorf("expected Held x3 after reopening, got: %v x%d", again.Status, again.Quantity)
	}
}

// productIDs se queda solo con los IDs, para imprimirlos en el mensaje de fallo.
func productIDs(ps []*Product) []string {
	out := make([]string, 0, len(ps))
	for _, p := range ps {
		out = append(out, p.ID)
	}
	return out
}

// ids se queda solo con los IDs de las reservas, para imprimirlos en el mensaje
// de fallo.
func ids(rs []*Reservation) []string {
	out := make([]string, 0, len(rs))
	for _, r := range rs {
		out = append(out, r.ID)
	}
	return out
}
