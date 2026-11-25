package reservations

// Bloque 1: seis consultas de práctica sobre las tablas products y
// reservations. Cada una es un texto SQL con nombre, igual que las de
// store.go. Los ? son los valores que cambian de una llamada a otra: los pone
// el test, en orden. Pruébalas antes a mano con cmd/sql:
//
//	go run ./cmd/sql "SELECT name, stock FROM products"
const (
	// Todos los productos, con sus tres columnas (id, name, stock), ordenados
	// por nombre.
	sqlPracticeProductsByName = `SELECT id, name, stock FROM products ORDER BY name`

	// El nombre de los productos con menos stock que el que se pasa, ordenados
	// por nombre.
	sqlPracticeLowStock = `SELECT name FROM products WHERE stock < ? ORDER BY name`

	// El id y la quantity de las reservas de un producto, de la más antigua a
	// la más nueva.
	sqlPracticeReservationsOf = `SELECT id, quantity FROM reservations WHERE product_id = ? ORDER BY id`

	// Cuántas reservas hay en un estado.
	sqlPracticeCountByStatus = `SELECT COUNT(*) FROM reservations WHERE status = ?`

	// Cambia el nombre de un producto. Primero llega el nombre nuevo y después
	// el id.
	sqlPracticeRename = `UPDATE products SET name = ? WHERE id = ?`

	// Borra las reservas que están en un estado.
	sqlPracticeDeleteByStatus = `DELETE FROM reservations WHERE status = ?`
)
