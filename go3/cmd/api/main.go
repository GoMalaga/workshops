// Bloque 2: la API de reservas como programa. Abre la base de datos, guarda los
// datos de ejemplo la primera vez y sirve los handlers de api.go en un puerto.
//
//	go run ./cmd/api                 # http://localhost:8080, fichero taquilla.db
//	go run ./cmd/api -addr :8081     # una segunda instancia en otro puerto, mismo fichero
//
// Pruébala desde otra terminal:
//
//	curl localhost:8080/products
//	curl -X POST localhost:8080/reserve -d '{"product_id":"p1","quantity":2}'
//	curl -X POST localhost:8080/reservations/5/confirm
package main

import (
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

	// Los datos de ejemplo, solo si el fichero es nuevo.
	if err := store.SeedDemo(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	log.Printf("taquilla listening on %s, data in %s", *addr, *dbPath)
	if err := http.ListenAndServe(*addr, reservations.NewHandler(store)); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
