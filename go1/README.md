# Go 1 — Fundamentos y lógica de dominio

Taller práctico de Go. Construyes el **motor de un sistema de reservas de stock**:
la lógica pura y sus tests, sin HTTP todavía.

> Invariante que gobierna todo: **el stock nunca baja de cero; nunca se reserva
> más de lo que hay.**

La guía de la sesión, las slides en PDF para leer con el repo al lado:
[guia.pdf](guia.pdf).

## Antes de empezar (Bloque 0)

Todos los comandos de este taller se ejecutan desde la carpeta `go1/` del repo.

```bash
./check-env.sh      # valida que tienes Go 1.22+
go run ./cmd/hello  # si imprime tu versión, estás listo
```

## Cómo se trabaja

Rellenas los `// TODO`. Los tests ya están escritos y **fallan**; tu trabajo
es hacer que **pasen**. Los tests que te damos no se tocan; los
que escribas tú, al final del mismo fichero.

### Bloque 1 — la máquina de estados (`reservation.go`)

El proyecto arranca con un hueco a propósito: `go test` no compila porque falta
un estado, y los tests saben cómo se llama. Añádelo al bloque de `iota` y haz
que `String()` lo imprima; `TestStatus_String` te dice cuándo lo tienes.

Después implementa `confirm()` y `cancel()`. Solo se puede confirmar o cancelar
una reserva que está `Held`; cualquier otra transición devuelve
`ErrIllegalTransition`.

```bash
go test -run 'TestStatus|TestConfirm|TestCancel' -v
```

Cuando pasen todos, falta un test: cancelar una reserva ya cancelada también es
ilegal, y ningún test lo comprueba. Escríbelo tú al final de `reservation_test.go`,
copiando el patrón de los demás, y haz que pase.

Has terminado el bloque cuando:

- [ ] Compila: `Canceled` está en el `iota` y en `String()`.
- [ ] `go test -run TestStatus -v` pasa.
- [ ] `go test -run 'TestConfirm|TestCancel' -v` pasa.
- [ ] Has escrito el sexto test, cancelar una reserva ya cancelada es ilegal, y pasa.

### Bloque 2 — el almacén (`store.go`)

Implementa `Reserve()` y `Cancel()`. `Confirm()` te lo damos resuelto como
ejemplo del patrón: buscar, comprobar nil, delegar. Después, `Reservations()`:
las reservas de un producto, ordenadas por ID. Un map no tiene orden, así que
recorre, acumula en un slice y ordena. Si el test pasa unas veces sí y otras no,
ya sabes por qué.

```bash
go test -run 'TestReserv|TestStore' -v
```

Has terminado el bloque cuando:

- [ ] `go test -run TestReserve_ -v` pasa: los cinco de `Reserve`.
- [ ] `go test -run TestStore -v` pasa: `Cancel` devuelve el stock, y solo una vez.
- [ ] `go test -run TestReservations -count=5 -v` pasa las cinco veces.

### Bloque 3 — la taquilla, de cero (`cmd/taquilla`)

Hasta aquí has hecho pasar tests sobre código ajeno. Ahora el programa lo escribes tú,
sin esqueleto ni tests: una taquilla por consola que lee órdenes de la entrada
estándar y las ejecuta contra tu `Store`.

Crea `cmd/taquilla/main.go`. Al arrancar, el programa da de alta un producto
`p1` con 10 unidades. Después lee una orden por línea y responde con una línea:

| Orden                           | Respuesta                                        |
|---------------------------------|--------------------------------------------------|
| `reserve <producto> <cantidad>` | el ID y el estado de la reserva nueva: `r1 Held` |
| `confirm <reserva>`             | `ok`                                             |
| `cancel <reserva>`              | `ok`                                             |
| `stock <producto>`              | las unidades disponibles: `7`                    |
| `quit` (o fin de la entrada)    | termina                                          |

Cualquier fallo se imprime como `error: <mensaje>` y el programa sigue leyendo.
Esta sesión es tu test: si tu programa responde exactamente esto, funciona.

```
$ go run ./cmd/taquilla
reserve p1 3
r1 Held
stock p1
7
confirm r1
ok
cancel r1
error: illegal state transition
reserve p1 8
error: out of stock
stock p9
error: product not found
quit
```

Para repetirla sin teclear:

```bash
printf 'reserve p1 3\nstock p1\nconfirm r1\ncancel r1\nreserve p1 8\nstock p9\nquit\n' | go run ./cmd/taquilla
```

Pistas, solo nombres; el resto está en [pkg.go.dev](https://pkg.go.dev):

- Tu paquete se importa como `reservations`, el nombre del módulo en `go.mod`.
- `bufio.NewScanner(os.Stdin)` lee línea a línea. `strings.Fields` parte una
  línea en palabras. `strconv.Atoi` convierte una palabra en número.
- `stock` necesita algo que el paquete todavía no ofrece desde fuera. Añádelo
  en `store.go`, y fíjate en por qué su nombre tiene que empezar por mayúscula.

Cuando funcione, teclea `reserve p1 -3` y mira el stock. Dieciocho tests que pasan y
la invariante rota: decide la regla, escribe tu primer test en un `_test.go`
nuevo y haz que pase.

Si te sobra tiempo, sin contrato ni tests: una orden `list <producto>` que
imprima lo que devuelve `Reservations()`, una reserva por línea. Y si `execute`
necesita el `Store` y la salida en varias funciones, un struct propio con un
método es la forma Go de agruparlos.

Has terminado el bloque cuando:

- [ ] `go run ./cmd/taquilla` responde la sesión de arriba línea a línea.
- [ ] `reserve p1 -3` ya no sube el stock: tu regla y tu test, en un `_test.go` nuevo.
- [ ] `go test ./...` pasa entero.

### Todo junto

```bash
go test ./...        # toda la suite pasa = has terminado
go test -race ./...  # (pista para la semana que viene...)
```

## Checkpoints (por si te atascas)

Si te quedas descolgado, puedes saltar a un punto de rescate. Los tags se
publican a lo largo de la sesión, así que haz `git fetch --tags` antes:

- `git checkout go1-step0`: punto de partida (todo por hacer)
- `git checkout go1-step1`: Bloque 1 ya resuelto (empieza el Bloque 2 desde aquí)
- `git checkout go1-step2`: Bloques 1 y 2 resueltos (empieza el Bloque 3 desde aquí)
- `git checkout go1-final`: todo resuelto, Bloque 3 incluido (referencia / solución)

Vuelve a tu trabajo con `git checkout -` (o el nombre de tu rama).
