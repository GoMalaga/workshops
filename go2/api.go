package reservations

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// NewHandler conecta las rutas con un Store y devuelve algo que un http.Server
// sabe servir. Patrones de Go 1.22: el método y la ruta en un solo texto, y
// comodines {id} que el handler lee con r.PathValue("id").
func NewHandler(store *Store) http.Handler {
	a := &api{store: store}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /products", a.products)
	mux.HandleFunc("GET /products/{id}", a.product)
	mux.HandleFunc("POST /products/{id}", a.addProduct)
	mux.HandleFunc("POST /reserve", a.reserve)
	mux.HandleFunc("POST /reservations/{id}/confirm", a.confirm)
	mux.HandleFunc("POST /reservations/{id}/cancel", a.cancel)
	return mux
}

// api guarda lo que necesitan los handlers. Cada handler es un método suyo,
// para llegar al Store sin una variable global.
type api struct {
	store *Store
}

// reserveRequest es el cuerpo de POST /reserve:
// {"product_id": "p1", "quantity": 2}.
type reserveRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

// GET /products -> 200 y el catálogo como un array JSON.
// El que escribimos juntos. Modelo: product, sin entrada que leer.
func (a *api) products(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, errorBody("TODO: GET /products"))
}

// GET /products/{id} -> 200 y el producto, o 404.
// EJEMPLO RESUELTO de un handler: lee la entrada (aquí, la ruta), llama al
// Store, le pasa cualquier error a writeError y escribe JSON. El handler viene
// escrito; writeError, más abajo, es tuyo: hasta entonces responde 500 a todo.
func (a *api) product(w http.ResponseWriter, r *http.Request) {
	p, err := a.store.Product(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

// POST /products/{id} -> 201 y el producto nuevo. El nombre y el stock llegan
// como campos de formulario: name=Ticket&stock=10.
// Viene escrito y no es de ningún paso: lo usa cmd/oversell, en el bloque 3,
// para crear un producto nuevo en cada carga.
func (a *api) addProduct(w http.ResponseWriter, r *http.Request) {
	stock, err := strconv.Atoi(r.FormValue("stock"))
	if err != nil || stock < 0 {
		writeJSON(w, http.StatusBadRequest, errorBody("stock must be a number, zero or more"))
		return
	}
	p := &Product{ID: r.PathValue("id"), Name: r.FormValue("name"), Stock: stock}
	if err := a.store.AddProduct(p); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

// POST /reserve -> 201 y la reserva nueva.
// El cuerpo llega como un reserveRequest. Si no es JSON válido es culpa del
// cliente: 400 y un mensaje, sin llegar al Store. Los errores del Store se
// convierten en códigos en writeError (404, 409, 400).
func (a *api) reserve(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, errorBody("TODO: POST /reserve"))
}

// POST /reservations/{id}/confirm -> 200 y la reserva como queda.
// El id viene en la ruta, como en product. Acaba en respondWithReservation,
// que escribes más abajo.
func (a *api) confirm(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, errorBody("TODO: POST /reservations/{id}/confirm"))
}

// POST /reservations/{id}/cancel -> 200 y la reserva como queda.
// La misma forma que confirm, con el otro método del Store.
func (a *api) cancel(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, errorBody("TODO: POST /reservations/{id}/cancel"))
}

// respondWithReservation vuelve a leer una reserva y la escribe con un 200.
// Confirm y Cancel no devuelven la reserva, solo un error, así que el handler
// la lee otra vez. confirm y cancel acaban los dos aquí.
func (a *api) respondWithReservation(w http.ResponseWriter, id string) {
	writeJSON(w, http.StatusNotImplemented, errorBody("TODO: respondWithReservation"))
}

// writeJSON pone el tipo de contenido y el código de estado, y escribe v como
// cuerpo. El orden importa: primero las cabeceras, después WriteHeader y al
// final el cuerpo.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError convierte un error del dominio en un código HTTP y un cuerpo JSON.
// Lo que no existe es 404, una regla del dominio que dice que no es 409
// (conflicto), una entrada mal formada es 400 y todo lo demás es 500: es culpa
// nuestra, no del cliente.
// Los errores son los cinco de reservation.go; errors.Is(err, ErrOutOfStock)
// dice cuál tienes (añade "errors" a los import de arriba: Go no acepta un
// import sin usar). Todos los handlers acaban aquí cuando algo va mal, así
// que la tabla está en un solo sitio.
func writeError(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusInternalServerError, errorBody("TODO: writeError: "+err.Error()))
}

// errorBody es la forma de todas las respuestas de error:
// {"error": "out of stock"}.
func errorBody(msg string) map[string]string {
	return map[string]string{"error": msg}
}
