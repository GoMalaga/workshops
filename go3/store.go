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
// driver los rellena con los argumentos que le pasas, nunca pegando texto.
const (
	sqlInsertProduct      = `INSERT INTO products (id, name, stock) VALUES (?, ?, ?)`
	sqlSelectProducts     = `SELECT id, name, stock FROM products ORDER BY id`
	sqlSelectProduct      = `SELECT id, name, stock FROM products WHERE id = ?`
	sqlSelectStock        = `SELECT stock FROM products WHERE id = ?`
	sqlUpdateStock        = `UPDATE products SET stock = stock + ? WHERE id = ?`
	sqlInsertReservation  = `INSERT INTO reservations (product_id, quantity, status) VALUES (?, ?, ?)`
	sqlSelectReservation  = `SELECT id, product_id, quantity, status FROM reservations WHERE id = ?`
	sqlUpdateStatus       = `UPDATE reservations SET status = ? WHERE id = ?`
	sqlSelectReservations = `SELECT id, product_id, quantity, status FROM reservations WHERE product_id = ? ORDER BY id`
	sqlCountConfirmed     = `SELECT COUNT(*) FROM reservations WHERE product_id = ? AND status = ?`
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
// escritura (lo explica la guía del bloque 3 de la sesión 2).
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

// AddProduct da de alta un producto: un INSERT. Exec es para las sentencias que
// no devuelven filas. Un segundo producto con el mismo ID falla: id es la
// PRIMARY KEY.
func (s *Store) AddProduct(p *Product) error {
	_, err := s.db.Exec(sqlInsertProduct, p.ID, p.Name, p.Stock)
	return err
}

// Products lista el catálogo, ordenado por ID. Query devuelve las filas; se
// recorren con Next, cada una se lee con Scan, y al final se pregunta rows.Err.
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
func (s *Store) Product(productID string) (*Product, error) {
	p := &Product{}
	err := s.db.QueryRow(sqlSelectProduct, productID).Scan(&p.ID, &p.Name, &p.Stock)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, err
	}
	return p, nil
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
func (s *Store) Reservation(reservationID string) (*Reservation, error) {
	return scanReservation(s.db.QueryRow(sqlSelectReservation, reservationID))
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
// tiene la misma forma más una escritura: el stock vuelve al producto.
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
//
// Sesión 2, bloque 3: la comprobación (SELECT) y la escritura (UPDATE) van dentro de una
// misma transacción. Con _txlock=immediate la transacción coge el bloqueo de
// escritura en Begin, así que un segundo Reserve espera a que este haga Commit
// y entonces lee el stock real. Sin la transacción, dos llamadas simultáneas
// leen 1 las dos y restan las dos.
func (s *Store) Reserve(productID string, qty int) (*Reservation, error) {
	if qty <= 0 {
		return nil, ErrInvalidQuantity
	}
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // no hace nada tras un Commit; si no lo hubo, lo deshace todo

	var stock int
	err = tx.QueryRow(sqlSelectStock, productID).Scan(&stock)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, err
	}
	if stock < qty {
		return nil, ErrOutOfStock
	}
	if _, err := tx.Exec(sqlUpdateStock, -qty, productID); err != nil {
		return nil, err
	}
	res, err := tx.Exec(sqlInsertReservation, productID, qty, Held)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &Reservation{ID: strconv.FormatInt(id, 10), ProductID: productID, Quantity: qty, Status: Held}, nil
}

// Cancel libera una reserva Held y DEVUELVE su cantidad al producto.
// El stock solo se devuelve si la transición de cancelar era legal.
// ErrReservationNotFound si la reserva no existe.
//
// Sesión 2, bloque 3: la misma transacción que en Reserve. Sin ella, dos Cancel
// simultáneos de la misma reserva verían Held los dos y devolverían el stock
// dos veces.
func (s *Store) Cancel(reservationID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	r, err := scanReservation(tx.QueryRow(sqlSelectReservation, reservationID))
	if err != nil {
		return err
	}
	if err := r.cancel(); err != nil {
		return err
	}
	if _, err := tx.Exec(sqlUpdateStatus, r.Status, reservationID); err != nil {
		return err
	}
	if _, err := tx.Exec(sqlUpdateStock, r.Quantity, r.ProductID); err != nil {
		return err
	}
	return tx.Commit()
}

// Reservations devuelve todas las reservas de un producto, ordenadas por ID.
// El orden lo pone la base de datos (ORDER BY): ni mapa que recorrer ni sort
// que llamar. Si el producto no existe o no tiene reservas: un slice vacío (no
// nil), nunca un error.
func (s *Store) Reservations(productID string) ([]*Reservation, error) {
	rows, err := s.db.Query(sqlSelectReservations, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rs := make([]*Reservation, 0)
	for rows.Next() {
		r := &Reservation{}
		var id int64
		if err := rows.Scan(&id, &r.ProductID, &r.Quantity, &r.Status); err != nil {
			return nil, err
		}
		r.ID = strconv.FormatInt(id, 10)
		rs = append(rs, r)
	}
	return rs, rows.Err()
}

// CountConfirmed dice cuántas reservas confirmadas tiene un producto.
// ErrProductNotFound si el producto no existe: COUNT devuelve 0 tanto si no hay
// reservas como si no hay producto, así que antes se comprueba con Stock.
// Ejercicio para casa: el método que se escribe de cero.
func (s *Store) CountConfirmed(productID string) (int, error) {
	if _, err := s.Stock(productID); err != nil {
		return 0, err
	}
	var n int
	err := s.db.QueryRow(sqlCountConfirmed, productID, Confirmed).Scan(&n)
	return n, err
}
