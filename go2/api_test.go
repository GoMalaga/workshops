package reservations

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Tests del bloque 2. httptest.NewRecorder es un ResponseWriter de mentira: el
// handler escribe en él y el test lee después el código de estado y el
// cuerpo. Sin puerto, sin red y sin arrancar ningún servidor.

// newAPI monta el handler sobre un Store nuevo con un producto: p1, stock 5.
// Devuelve también el Store, para los tests que preparan o comprueban datos
// directamente.
func newAPI(t *testing.T) (http.Handler, *Store) {
	t.Helper()
	s := newStoreWithStock(t, 5)
	return NewHandler(s), s
}

// do manda una petición al handler y devuelve lo que ha respondido.
func do(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// decode lee un cuerpo JSON en v y hace fallar el test si no es JSON.
func decode(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.Unmarshal(rec.Body.Bytes(), v); err != nil {
		t.Fatalf("body is not JSON: %v\n%s", err, rec.Body.String())
	}
}

// Paso 1. GET /products responde 200 con un array JSON que tiene el único
// producto que hay, p1 con stock 5.
func TestAPI_products_listsCatalog(t *testing.T) {
	h, _ := newAPI(t)
	rec := do(t, h, "GET", "/products", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var ps []Product
	decode(t, rec, &ps)
	if len(ps) != 1 || ps[0].ID != "p1" || ps[0].Stock != 5 {
		t.Errorf("expected [p1 with stock 5], got: %+v", ps)
	}
}

// Paso 2. GET /products/ghost usa el handler que viene dado, así que lo único
// que se prueba es writeError: ErrProductNotFound tiene que acabar en un 404.
func TestAPI_product_unknown_is404(t *testing.T) {
	h, _ := newAPI(t)
	rec := do(t, h, "GET", "/products/ghost", "")
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

// Paso 3, el caso normal. POST /reserve responde 201 con la reserva en JSON:
// tiene un ID, el estado viaja con su nombre ("Held", no un número), y un GET
// del producto después muestra que el stock ha pasado de 5 a 3.
func TestAPI_reserve_created(t *testing.T) {
	h, _ := newAPI(t)
	rec := do(t, h, "POST", "/reserve", `{"product_id": "p1", "quantity": 2}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var res Reservation
	decode(t, rec, &res)
	if res.ID == "" || res.ProductID != "p1" || res.Quantity != 2 {
		t.Errorf("expected a reservation of 2 x p1 with an ID, got: %+v", res)
	}
	if !strings.Contains(rec.Body.String(), `"status":"Held"`) {
		t.Errorf("expected the status to travel as its name, got: %s", rec.Body.String())
	}
	stock := do(t, h, "GET", "/products/p1", "")
	if !strings.Contains(stock.Body.String(), `"stock":3`) {
		t.Errorf("expected stock 3 after reserving 2, got: %s", stock.Body.String())
	}
}

// Pedir 6 cuando hay 5 es un conflicto con el estado actual: 409.
func TestAPI_reserve_outOfStock_is409(t *testing.T) {
	h, _ := newAPI(t)
	rec := do(t, h, "POST", "/reserve", `{"product_id": "p1", "quantity": 6}`)
	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

// Reservar un producto que no existe es un 404.
func TestAPI_reserve_unknownProduct_is404(t *testing.T) {
	h, _ := newAPI(t)
	rec := do(t, h, "POST", "/reserve", `{"product_id": "ghost", "quantity": 1}`)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

// El JSON está bien formado pero la cantidad es 0: el Store la rechaza con
// ErrInvalidQuantity y writeError lo convierte en un 400.
func TestAPI_reserve_invalidQuantity_is400(t *testing.T) {
	h, _ := newAPI(t)
	rec := do(t, h, "POST", "/reserve", `{"product_id": "p1", "quantity": 0}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

// El cuerpo está cortado por la mitad, así que no llega al Store: el propio
// handler responde 400, y el cuerpo sigue siendo JSON con un campo "error".
func TestAPI_reserve_badJSON_is400(t *testing.T) {
	h, _ := newAPI(t)
	rec := do(t, h, "POST", "/reserve", `{"product_id": "p1", "quantity": `)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	var body map[string]string
	decode(t, rec, &body)
	if body["error"] == "" {
		t.Errorf(`expected {"error": "..."} in the body, got: %s`, rec.Body.String())
	}
}

// Paso 4. La reserva se crea directamente en el Store; lo que se prueba es el
// POST que la confirma: 200, y el cuerpo es la reserva, ya Confirmed.
func TestAPI_confirm_ok(t *testing.T) {
	h, s := newAPI(t)
	r := mustReserve(t, s, "p1", 1)
	rec := do(t, h, "POST", "/reservations/"+r.ID+"/confirm", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"status":"Confirmed"`) {
		t.Errorf("expected Confirmed in the body, got: %s", rec.Body.String())
	}
}

// Confirmar dos veces: la segunda es una transición ilegal, así que 409.
func TestAPI_confirm_twice_is409(t *testing.T) {
	h, s := newAPI(t)
	r := mustReserve(t, s, "p1", 1)
	do(t, h, "POST", "/reservations/"+r.ID+"/confirm", "")
	rec := do(t, h, "POST", "/reservations/"+r.ID+"/confirm", "")
	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409 on the second confirm, got %d: %s", rec.Code, rec.Body.String())
	}
}

// Confirmar una reserva que no existe es un 404.
func TestAPI_confirm_unknown_is404(t *testing.T) {
	h, _ := newAPI(t)
	rec := do(t, h, "POST", "/reservations/999/confirm", "")
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

// Paso 5. Cancelar responde 200 con la reserva como Canceled, y el stock,
// leído directamente del Store, vuelve a 5.
func TestAPI_cancel_returnsStock(t *testing.T) {
	h, s := newAPI(t)
	r := mustReserve(t, s, "p1", 2)
	rec := do(t, h, "POST", "/reservations/"+r.ID+"/cancel", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"status":"Canceled"`) {
		t.Errorf("expected Canceled in the body, got: %s", rec.Body.String())
	}
	if got := stockOf(t, s, "p1"); got != 5 {
		t.Errorf("expected stock back at 5, got: %d", got)
	}
}

// Cancelar una reserva que no existe es un 404.
func TestAPI_cancel_unknown_is404(t *testing.T) {
	h, _ := newAPI(t)
	rec := do(t, h, "POST", "/reservations/999/cancel", "")
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

// POST /products/{id} viene resuelto y no es de ningún paso: lo usa
// cmd/oversell en el bloque 3. Crea el producto con el nombre y el stock del
// formulario, responde 201, y un GET después lo encuentra.
func TestAPI_addProduct_created(t *testing.T) {
	h, _ := newAPI(t)
	rec := do(t, h, "POST", "/products/p9?name=Vinyl&stock=7", "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var p Product
	decode(t, do(t, h, "GET", "/products/p9", ""), &p)
	if p.ID != "p9" || p.Name != "Vinyl" || p.Stock != 7 {
		t.Errorf("expected p9 Vinyl with stock 7, got: %+v", p)
	}
}

// Un stock que no es un número no llega al Store: 400.
func TestAPI_addProduct_badStock_is400(t *testing.T) {
	h, _ := newAPI(t)
	rec := do(t, h, "POST", "/products/p9?name=Vinyl&stock=many", "")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}
