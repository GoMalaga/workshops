package reservations

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// Tests del bloque 1. Cada test lanza una de las consultas de consultas.go
// sobre los datos de ejemplo (demo.go) y compara las filas que devuelve.

// newDemoStore abre un Store nuevo con los datos de ejemplo.
func newDemoStore(t *testing.T) *Store {
	t.Helper()
	s, _ := newStore(t)
	if err := s.SeedDemo(); err != nil {
		t.Fatalf("seed demo data: %v", err)
	}
	return s
}

// rowsOf lanza un SELECT y devuelve cada fila como un texto, con las columnas
// separadas por comas: "p1,Ticket,6".
func rowsOf(t *testing.T, s *Store, query string, args ...any) []string {
	t.Helper()
	rows, err := s.db.Query(query, args...)
	if err != nil {
		t.Fatalf("query failed: %v\n%s", err, query)
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		t.Fatalf("columns: %v", err)
	}
	out := []string{}
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			t.Fatalf("scan: %v", err)
		}
		parts := make([]string, len(vals))
		for i, v := range vals {
			if b, ok := v.([]byte); ok {
				v = string(b)
			}
			parts[i] = fmt.Sprint(v)
		}
		out = append(out, strings.Join(parts, ","))
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows: %v", err)
	}
	return out
}

// expectRows compara las filas obtenidas con las esperadas, en orden.
func expectRows(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected rows %q, got %q", want, got)
	}
}

// Paso 1. Los tres productos con sus tres columnas, por orden alfabético del
// nombre: Poster, Ticket, Vinyl.
func TestPractice_productsByName(t *testing.T) {
	s := newDemoStore(t)
	got := rowsOf(t, s, sqlPracticeProductsByName)
	expectRows(t, got, []string{"p2,Poster,3", "p1,Ticket,6", "p3,Vinyl,0"})
}

// Paso 2. Con menos de 5 unidades están Poster (3) y Vinyl (0). Ticket tiene 6
// y no sale. Solo se pide la columna name.
func TestPractice_lowStock(t *testing.T) {
	s := newDemoStore(t)
	got := rowsOf(t, s, sqlPracticeLowStock, 5)
	expectRows(t, got, []string{"Poster", "Vinyl"})
}

// Paso 3. El producto p1 tiene las reservas 1, 2 y 4. Se piden dos columnas,
// id y quantity, de la más antigua a la más nueva.
func TestPractice_reservationsOf(t *testing.T) {
	s := newDemoStore(t)
	got := rowsOf(t, s, sqlPracticeReservationsOf, "p1")
	expectRows(t, got, []string{"1,2", "2,1", "4,1"})
}

// Paso 4. Hay dos reservas Confirmed y una Held. La consulta devuelve una sola
// fila con una sola columna: el número.
func TestPractice_countByStatus(t *testing.T) {
	s := newDemoStore(t)
	expectRows(t, rowsOf(t, s, sqlPracticeCountByStatus, Confirmed), []string{"2"})
	expectRows(t, rowsOf(t, s, sqlPracticeCountByStatus, Held), []string{"1"})
}

// Paso 5. Cambiar el nombre de p2 no toca ni su stock ni los otros productos.
func TestPractice_rename(t *testing.T) {
	s := newDemoStore(t)
	if _, err := s.db.Exec(sqlPracticeRename, "Sticker", "p2"); err != nil {
		t.Fatalf("exec failed: %v\n%s", err, sqlPracticeRename)
	}
	got := rowsOf(t, s, `SELECT id, name, stock FROM products ORDER BY id`)
	expectRows(t, got, []string{"p1,Ticket,6", "p2,Sticker,3", "p3,Vinyl,0"})
}

// Paso 6. Borrar las reservas Canceled quita solo la número 3. Las otras tres
// siguen ahí.
func TestPractice_deleteByStatus(t *testing.T) {
	s := newDemoStore(t)
	if _, err := s.db.Exec(sqlPracticeDeleteByStatus, Canceled); err != nil {
		t.Fatalf("exec failed: %v\n%s", err, sqlPracticeDeleteByStatus)
	}
	got := rowsOf(t, s, `SELECT id FROM reservations ORDER BY id`)
	expectRows(t, got, []string{"1", "2", "4"})
}
