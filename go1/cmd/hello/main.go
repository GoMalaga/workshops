package main

import (
	"fmt"
	"runtime"
)

// Bloque 0: si esto imprime tu versión de Go, tu entorno está listo.
// Ejecútalo con:  go run ./cmd/hello
func main() {
	fmt.Printf("Go funciona. Versión: %s\n", runtime.Version())
}
