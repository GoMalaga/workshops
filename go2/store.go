package reservations

import (
	"database/sql"
	"errors"
	"strconv"

	_ "modernc.org/sqlite" // registra el driver "sqlite" en database/sql
)

// schema crea las dos tablas la primera vez que se abre el fichero. Con IF NOT
// EXISTS se puede lanzar en cada arranque. El estado se guarda como el número
// que hay detrás del iota.
const schema = `
CREATE TABLE IF NOT EXISTS products (
	id    TEXT PRIMARY KEY,
	name  TEXT NOT NULL,
	stock INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS reservations (
	id         INTEGER PRIMARY KEY,
	product_id TEXT NOT NULL REFERENCES products(id),
	quantity   INTEGER NOT NULL,
	status     INTEGER NOT NULL
);`

// Las consultas, escritas una sola vez y con nombre. Los ? son marcadores: el
// driver los rellena con los argumentos que le pasas, en orden, y nunca
// pegando texto. Hay una escrita, sqlSelectStock, la que usa Stock: es tu
// ejemplo. Las otras ocho son tuyas. El comentario dice qué tiene que hacer
// cada una, el schema de arriba da los nombres de las tablas y las columnas, y
// el Scan del método que la usa dice qué columnas devolver y en qué orden.
const (
	// TODO: inserta un producto con su id, su name y su stock.
	sqlInsertProduct = `TODO`

	// TODO: todos los productos, ordenados por id. Mira el Scan de Products.
	sqlSelectProducts = `TODO`

	// TODO: un producto por su id.
	sqlSelectProduct = `TODO`

	// El stock de un producto. Viene escrita: es el ejemplo de consulta.
	sqlSelectStock = `SELECT stock FROM products WHERE id = ?`

	// TODO: suma una cantidad al stock de un producto: negativa para reservar,
	// positiva para devolver. La misma consulta sirve para Reserve y para Cancel.
	sqlUpdateStock = `TODO`

	// TODO: inserta una reserva con su product_id, su quantity y su status.
	// El id no: lo elige la base de datos (INTEGER PRIMARY KEY).
	sqlInsertReservation = `TODO`

	// TODO: una reserva por su id. Mira en scanReservation qué columnas lee y
	// en qué orden.
	sqlSelectReservation = `TODO`

	// TODO: cambia el status de una reserva, por su id. Mira en Confirm en qué
	// orden le pasa los valores.
	sqlUpdateStatus = `TODO`

	// TODO: todas las reservas de un producto, la más antigua primero. Las mismas
	// columnas y en el mismo orden que sqlSelectReservation, para que Scan las
	// lea igual.
	sqlSelectReservations = `TODO`
)

// Store es el Store de la sesión 1 con los mapas cambiados por un fichero
// SQLite. Las mismas operaciones, los mismos errores y el mismo invariante; los
// datos sobreviven a un reinicio.
type Store struct {
	db *sql.DB
}

// OpenStore abre (o crea) el fichero de la base de datos y se asegura de que
// las tablas existen. busy_timeout: quien va a escribir y encuentra el fichero
// bloqueado espera hasta 5 s en vez de fallar. _txlock=immediate: cada
// transacción coge el bloqueo de escritura al empezar, no en su primera
// escritura (el bloque 3 vuelve sobre esto).
func OpenStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_txlock=immediate")
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

// Close libera el fichero. Se llama al acabar el programa.
func (s *Store) Close() error {
	return s.db.Close()
}

// AddProduct da de alta un producto.
// Modelo: la última línea de Confirm, más abajo, que también escribe y solo
// mira el error.
func (s *Store) AddProduct(p *Product) error {
	return errors.New("TODO: AddProduct")
}

// Products lista el catálogo, ordenado por ID.
// EJEMPLO RESUELTO de leer muchas filas: Query devuelve las filas; se recorren
// con Next, cada una se lee con Scan, y al final se pregunta rows.Err. El
// método viene escrito; su consulta, sqlSelectProducts, es tuya.
func (s *Store) Products() ([]*Product, error) {
	rows, err := s.db.Query(sqlSelectProducts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ps := make([]*Product, 0)
	for rows.Next() {
		p := &Product{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Stock); err != nil {
			return nil, err
		}
		ps = append(ps, p)
	}
	return ps, rows.Err()
}

// Product lee un producto. ErrProductNotFound si no hay esa fila.
// Modelo: Stock, más abajo, con tres columnas en vez de una.
func (s *Store) Product(productID string) (*Product, error) {
	return nil, errors.New("TODO: Product")
}

// Stock dice cuántas unidades quedan por reservar.
// EJEMPLO RESUELTO de leer un valor: QueryRow lanza el SELECT y Scan copia la
// columna en la variable. Si no hay ninguna fila, el error es sql.ErrNoRows,
// que aquí quiere decir "no existe el producto": se convierte en
// ErrProductNotFound. Cualquier otro error (el fichero no está, el disco está
// lleno) se devuelve tal cual.
func (s *Store) Stock(productID string) (int, error) {
	var stock int
	err := s.db.QueryRow(sqlSelectStock, productID).Scan(&stock)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrProductNotFound
	}
	if err != nil {
		return 0, err
	}
	return stock, nil
}

// Reservation lee una reserva. ErrReservationNotFound si no existe.
// El ID es el INTEGER PRIMARY KEY de la fila, llevado como texto ("1", "2"...).
// scanReservation, más abajo, hace el Scan y convierte el "no hay fila" en
// ErrReservationNotFound.
func (s *Store) Reservation(reservationID string) (*Reservation, error) {
	return nil, errors.New("TODO: Reservation")
}

// scanReservation convierte una fila de la tabla reservations en una
// Reservation. Sirve igual para un QueryRow suelto que para uno dentro de una
// transacción.
func scanReservation(row *sql.Row) (*Reservation, error) {
	r := &Reservation{}
	var id int64
	err := row.Scan(&id, &r.ProductID, &r.Quantity, &r.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrReservationNotFound
	}
	if err != nil {
		return nil, err
	}
	r.ID = strconv.FormatInt(id, 10)
	return r, nil
}

// Confirm carga la reserva, deja decidir al dominio (confirm() conoce las
// transiciones legales) y guarda el estado nuevo.
// EJEMPLO RESUELTO de la forma cargar -> método del dominio -> guardar. Cancel
// tiene la misma forma más una escritura: el stock vuelve al producto. El
// método viene escrito; su consulta, sqlUpdateStatus, es tuya.
func (s *Store) Confirm(reservationID string) error {
	r, err := s.Reservation(reservationID)
	if err != nil {
		return err
	}
	if err := r.confirm(); err != nil {
		return err
	}
	_, err = s.db.Exec(sqlUpdateStatus, r.Status, reservationID)
	return err
}

// Reserve comprueba el stock, lo resta y registra una reserva Held.
// INVARIANTE: el stock nunca baja de cero.
// ErrInvalidQuantity si qty es cero o negativa; ErrProductNotFound si el
// producto no existe; ErrOutOfStock si no hay suficiente.
// Modelos: Stock para la lectura y la última línea de Confirm para las
// escrituras. El Result del INSERT dice qué ID eligió la base de datos:
// res.LastInsertId(). Llévalo como texto (strconv).
func (s *Store) Reserve(productID string, qty int) (*Reservation, error) {
	return nil, errors.New("TODO: Reserve")
}

// Cancel libera una reserva Held y DEVUELVE su cantidad al producto.
// El stock solo se devuelve si la transición de cancelar era legal.
// ErrReservationNotFound si la reserva no existe.
// Modelo: Confirm. No hace falta ninguna consulta nueva.
func (s *Store) Cancel(reservationID string) error {
	return errors.New("TODO: Cancel")
}

// Reservations devuelve todas las reservas de un producto, ordenadas por ID.
// El orden lo pone la base de datos (ORDER BY): ni mapa que recorrer ni sort
// que llamar. Si el producto no existe o no tiene reservas: un slice vacío (no
// nil), nunca un error.
// Modelo: Products, más arriba.
func (s *Store) Reservations(productID string) ([]*Reservation, error) {
	return nil, errors.New("TODO: Reservations")
}
