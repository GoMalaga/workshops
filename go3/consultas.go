package reservations

// Bloque 1: seis consultas de práctica sobre las tablas products y
// reservations. Cada una es un texto SQL con nombre, igual que las de
// store.go, que vienen escritas y te sirven de ejemplo. El comentario dice
// qué tiene que devolver o hacer cada una: cambia el TODO por la sentencia.
//
// Los ? son los valores que cambian de una llamada a otra: los pone el test,
// en orden. Pruébalas antes a mano con cmd/sql, con un valor en vez del ?:
//
//	go run ./cmd/sql "SELECT name, stock FROM products"
const (
	// TODO: todos los productos, con sus tres columnas (id, name, stock),
	// ordenados por nombre.
	sqlPracticeProductsByName = `TODO`

	// TODO: el nombre de los productos con menos stock que el que se pasa,
	// ordenados por nombre.
	sqlPracticeLowStock = `TODO`

	// TODO: el id y la quantity de las reservas de un producto, de la más
	// antigua a la más nueva.
	sqlPracticeReservationsOf = `TODO`

	// TODO: cuántas reservas hay en un estado.
	sqlPracticeCountByStatus = `TODO`

	// TODO: cambia el nombre de un producto. Primero llega el nombre nuevo y
	// después el id.
	sqlPracticeRename = `TODO`

	// TODO: borra las reservas que están en un estado.
	sqlPracticeDeleteByStatus = `TODO`
)
