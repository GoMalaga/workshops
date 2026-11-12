package main

import (
	"database/sql"
	"fmt"
	"os"
	"runtime"

	_ "modernc.org/sqlite" // registra el driver "sqlite" en database/sql
)

// Bloque 0: comprueba que Go, el driver de SQLite y esta carpeta funcionan.
// Crea un fichero hello.db aquí mismo, una tabla, escribe una fila, la lee y
// borra el fichero. Si imprime "ok", el entorno está listo para la sesión.
// Ejecútalo con:  go run ./cmd/hello
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		fmt.Fprintln(os.Stderr, "Si habla de \"unable to open\", \"readonly\" o \"permission denied\", en esta carpeta no se puede escribir: copia el proyecto a otra y repite.")
		os.Exit(1)
	}
}

func run() error {
	const path = "hello.db"
	defer os.Remove(path)

	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=busy_timeout(5000)")
	if err != nil {
		return fmt.Errorf("abrir la base de datos: %w", err)
	}
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS hello (id INTEGER PRIMARY KEY, msg TEXT NOT NULL)`); err != nil {
		return fmt.Errorf("crear la tabla: %w", err)
	}
	if _, err := db.Exec(`INSERT INTO hello (msg) VALUES (?)`, "hola, 42"); err != nil {
		return fmt.Errorf("escribir una fila: %w", err)
	}
	var msg string
	if err := db.QueryRow(`SELECT msg FROM hello WHERE id = 1`).Scan(&msg); err != nil {
		return fmt.Errorf("leer la fila: %w", err)
	}
	fmt.Printf("ok. %s; SQLite crea, escribe y lee en esta carpeta: %q\n", runtime.Version(), msg)
	return nil
}
