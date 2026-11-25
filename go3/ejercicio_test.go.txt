package reservations

import (
	"errors"
	"net/http"
	"testing"
)

// Tests del ejercicio para casa: el endpoint que se escribe de cero. Usan los datos de
// ejemplo: p1 tiene dos reservas Confirmed y una Held; p2 tiene una Canceled.

// Paso 1. CountConfirmed cuenta solo las reservas Confirmed del producto: 2 en
// p1 (la Held no cuenta) y 0 en p2 (la suya está Canceled).
func TestCountConfirmed_countsOnlyConfirmed(t *testing.T) {
	s := newDemoStore(t)
	n, err := s.CountConfirmed("p1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 confirmed reservations for p1, got %d", n)
	}
	n, err = s.CountConfirmed("p2")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 confirmed reservations for p2, got %d", n)
	}
}

// Un producto que no existe no tiene "cero reservas": es un error,
// ErrProductNotFound.
func TestCountConfirmed_unknownProduct(t *testing.T) {
	s := newDemoStore(t)
	if _, err := s.CountConfirmed("ghost"); !errors.Is(err, ErrProductNotFound) {
		t.Errorf("expected ErrProductNotFound, got: %v", err)
	}
}

// Paso 2. GET /products/p1/confirmed responde 200 con un JSON que lleva el id
// del producto y el número: {"product_id":"p1","confirmed":2}.
func TestAPI_confirmedCount_ok(t *testing.T) {
	h := NewHandler(newDemoStore(t))
	rec := do(t, h, "GET", "/products/p1/confirmed", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		ProductID string `json:"product_id"`
		Confirmed int    `json:"confirmed"`
	}
	decode(t, rec, &body)
	if body.ProductID != "p1" || body.Confirmed != 2 {
		t.Errorf("expected product_id p1 and confirmed 2, got: %+v", body)
	}
}

// Con un producto que no existe responde 404, y el cuerpo es el JSON de error
// de writeError, no el texto que pone el mux cuando no encuentra la ruta.
func TestAPI_confirmedCount_unknown_is404(t *testing.T) {
	h := NewHandler(newDemoStore(t))
	rec := do(t, h, "GET", "/products/ghost/confirmed", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	var body map[string]string
	decode(t, rec, &body)
	if body["error"] == "" {
		t.Errorf("expected a JSON body with an \"error\" field, got: %s", rec.Body.String())
	}
}
