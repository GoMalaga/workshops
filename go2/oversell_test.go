package reservations

import (
	"sync"
	"sync/atomic"
	"testing"
)

// Bloque 3. Un producto, 10 unidades, y 100 clientes que piden 1 cada uno en el
// mismo instante. Tienen que entrar exactamente 10 y el stock tiene que acabar
// en 0, nunca por debajo. Lánzalo también con -race: no avisa de nada. Aquí no
// hay memoria de Go compartida; lo compartido es la fila de la base de datos,
// y eso -race no lo ve.
func TestConcurrentReserves_neverOversell(t *testing.T) {
	s := newStoreWithStock(t, 10)
	const clients = 100

	start := make(chan struct{}) // se cierra cuando todos están listos: es el pistoletazo de salida
	var wg sync.WaitGroup
	var succeeded atomic.Int32
	for i := 0; i < clients; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if _, err := s.Reserve("p1", 1); err == nil {
				succeeded.Add(1)
			}
		}()
	}
	close(start)
	wg.Wait()

	stock := stockOf(t, s, "p1")
	if stock < 0 {
		t.Fatalf("oversold: stock ended at %d and %d reservations succeeded; only 10 units existed", stock, succeeded.Load())
	}
	if succeeded.Load() != 10 || stock != 0 {
		t.Errorf("expected exactly 10 successful reservations and stock 0, got %d and %d", succeeded.Load(), stock)
	}
}

// Bloque 3, paso 3. 10 reservas de una unidad dejan el stock en 90, y cada
// reserva la cancelan 4 clientes en el mismo instante. De cada cuatro solo
// puede entrar una, así que entran 10 cancelaciones y el stock vuelve a 100,
// nunca por encima. Sin transacción, varios clientes leen la misma reserva
// todavía Held y cada uno devuelve su unidad por su cuenta.
func TestConcurrentCancels_returnStockOnce(t *testing.T) {
	s := newStoreWithStock(t, 100)
	const reservations, clients = 10, 4

	start := make(chan struct{})
	var wg sync.WaitGroup
	var succeeded atomic.Int32
	for i := 0; i < reservations; i++ {
		r := mustReserve(t, s, "p1", 1)
		for j := 0; j < clients; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				if err := s.Cancel(r.ID); err == nil {
					succeeded.Add(1)
				}
			}()
		}
	}
	close(start)
	wg.Wait()

	stock := stockOf(t, s, "p1")
	if stock > 100 {
		t.Fatalf("stock returned more than once: it ended at %d and %d cancellations succeeded; only 100 units existed", stock, succeeded.Load())
	}
	if succeeded.Load() != reservations || stock != 100 {
		t.Errorf("expected exactly %d successful cancellations and stock 100, got %d and %d", reservations, succeeded.Load(), stock)
	}
}
