package reservations

// Los datos de ejemplo de la sesión 3: tres productos y cuatro reservas, una
// en cada estado y dos confirmadas. Son los que ves con cmd/sql y los que
// sirve cmd/api la primera vez. El status es el número del estado: 0 es Held,
// 1 es Confirmed y 2 es Canceled.
const demoData = `
INSERT INTO products (id, name, stock) VALUES
	('p1', 'Ticket', 6),
	('p2', 'Poster', 3),
	('p3', 'Vinyl', 0);
INSERT INTO reservations (product_id, quantity, status) VALUES
	('p1', 2, 1),
	('p1', 1, 0),
	('p2', 1, 2),
	('p1', 1, 1);`

// SeedDemo guarda los datos de ejemplo si la base de datos está vacía. Si ya
// hay productos no toca nada, así que se puede llamar en cada arranque.
func (s *Store) SeedDemo() error {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM products`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	_, err := s.db.Exec(demoData)
	return err
}
