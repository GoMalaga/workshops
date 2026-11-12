// Bloque 2: la API de reservas como programa. Abre la base de datos, da de alta
// el catálogo la primera vez y sirve los handlers de api.go en un puerto.
//
//	go run ./cmd/api                 # http://localhost:8080, fichero taquilla.db
//	go run ./cmd/api -addr :8081     # una segunda instancia en otro puerto, mismo fichero
//
// Pruébala desde otra terminal:
//
//	curl localhost:8080/products
//	curl -X POST localhost:8080/reserve -d '{"product_id":"p1","quantity":2}'
//	curl -X POST localhost:8080/reservations/1/confirm
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"reservations"
)

func main() {
	addr := flag.String("addr", ":8080", "address to listen on")
	dbPath := flag.String("db", "taquilla.db", "SQLite file")
	flag.Parse()

	store, err := reservations.OpenStore(*dbPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	defer store.Close()

	if err := seed(store); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	log.Printf("taquilla listening on %s, data in %s", *addr, *dbPath)
	if err := http.ListenAndServe(*addr, reservations.NewHandler(store)); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// seed da de alta el catálogo si el fichero es nuevo. En un reinicio los
// productos ya están y no hace nada: para eso sirve el bloque 1.
func seed(store *reservations.Store) error {
	_, err := store.Stock("p1")
	if err == nil {
		return nil
	}
	if !errors.Is(err, reservations.ErrProductNotFound) {
		return err
	}
	for _, p := range []*reservations.Product{
		{ID: "p1", Name: "Ticket", Stock: 10},
		{ID: "p2", Name: "Poster", Stock: 3},
	} {
		if err := store.AddProduct(p); err != nil {
			return err
		}
	}
	return nil
}
