package reservations

import (
	"encoding/json"
	"errors"
	"net/http"
)

// NewHandler conecta las rutas con un Store y devuelve algo que un http.Server
// sabe servir. Patrones de Go 1.22: el método y la ruta en un solo texto, y
// comodines {id} que el handler lee con r.PathValue("id").
func NewHandler(store *Store) http.Handler {
	a := &api{store: store}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /products", a.products)
	mux.HandleFunc("GET /products/{id}", a.product)
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
func (a *api) products(w http.ResponseWriter, r *http.Request) {
	ps, err := a.store.Products()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ps)
}

// GET /products/{id} -> 200 y el producto, o 404.
// EJEMPLO RESUELTO de un handler: lee la entrada (aquí, la ruta), llama al
// Store, convierte el error en un código de estado y escribe JSON.
func (a *api) product(w http.ResponseWriter, r *http.Request) {
	p, err := a.store.Product(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

// POST /reserve -> 201 y la reserva nueva.
// Un cuerpo que no es JSON válido es culpa del cliente: 400 y un mensaje. Los
// errores del Store se convierten en códigos en writeError (404, 409, 400).
func (a *api) reserve(w http.ResponseWriter, r *http.Request) {
	var req reserveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody("invalid JSON: "+err.Error()))
		return
	}
	res, err := a.store.Reserve(req.ProductID, req.Quantity)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, res)
}

// POST /reservations/{id}/confirm -> 200 y la reserva como queda.
func (a *api) confirm(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := a.store.Confirm(id); err != nil {
		writeError(w, err)
		return
	}
	a.respondWithReservation(w, id)
}

// POST /reservations/{id}/cancel -> 200 y la reserva como queda.
func (a *api) cancel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := a.store.Cancel(id); err != nil {
		writeError(w, err)
		return
	}
	a.respondWithReservation(w, id)
}

// respondWithReservation vuelve a leer una reserva y la escribe con un 200.
func (a *api) respondWithReservation(w http.ResponseWriter, id string) {
	res, err := a.store.Reservation(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
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
func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ErrProductNotFound), errors.Is(err, ErrReservationNotFound):
		status = http.StatusNotFound
	case errors.Is(err, ErrOutOfStock), errors.Is(err, ErrIllegalTransition):
		status = http.StatusConflict
	case errors.Is(err, ErrInvalidQuantity):
		status = http.StatusBadRequest
	}
	writeJSON(w, status, errorBody(err.Error()))
}

// errorBody es la forma de todas las respuestas de error:
// {"error": "out of stock"}.
func errorBody(msg string) map[string]string {
	return map[string]string{"error": msg}
}
