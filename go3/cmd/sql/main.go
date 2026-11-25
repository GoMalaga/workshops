// Bloque 1: lanza una sentencia SQL sobre la base de datos de la taquilla y
// enseña el resultado. Si el fichero es nuevo, lo crea con las tablas y los
// datos de ejemplo.
//
//	go run ./cmd/sql "SELECT * FROM products"
//	go run ./cmd/sql "SELECT name, stock FROM products WHERE stock < 5"
//	go run ./cmd/sql "UPDATE products SET stock = 10 WHERE id = 'p3'"
//
// La sentencia va entre comillas dobles para que la terminal no toque el *.
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"reservations"
)

func main() {
	dbPath := flag.String("db", "taquilla.db", "SQLite file")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: sql [-db taquilla.db] 'SELECT * FROM products'")
		os.Exit(2)
	}
	if err := run(*dbPath, flag.Arg(0)); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(path, stmt string) error {
	// OpenStore crea las tablas si no existen; SeedDemo guarda los datos de
	// ejemplo si la base de datos está vacía.
	store, err := reservations.OpenStore(path)
	if err != nil {
		return err
	}
	if err := store.SeedDemo(); err != nil {
		store.Close()
		return err
	}
	store.Close()

	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)")
	if err != nil {
		return err
	}
	defer db.Close()

	// Un SELECT devuelve filas y se lanza con Query. Todo lo demás (INSERT,
	// UPDATE, DELETE) no devuelve filas y se lanza con Exec.
	if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(stmt)), "SELECT") {
		res, err := db.Exec(stmt)
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		fmt.Printf("ok, %d rows changed\n", n)
		return nil
	}

	rows, err := db.Query(stmt)
	if err != nil {
		return err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, strings.Join(cols, "\t"))
	count := 0
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return err
		}
		cells := make([]string, len(vals))
		for i, v := range vals {
			if b, ok := v.([]byte); ok {
				v = string(b)
			}
			cells[i] = fmt.Sprint(v)
		}
		fmt.Fprintln(w, strings.Join(cells, "\t"))
		count++
	}
	if err := rows.Err(); err != nil {
		return err
	}
	w.Flush()
	fmt.Printf("(%d rows)\n", count)
	return nil
}
