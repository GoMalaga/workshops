# Taller de Go · Sesión 2: persistencia y API

El motor de reservas de la sesión 1, guardado en SQLite y servido por HTTP.
Partes del árbol final de la sesión 1: todo lo de Go 1 está resuelto aquí.

## Las guías

Hay una guía por bloque. Cada una explica el código que viene hecho y después
da el paso a paso, con el test que comprueba cada paso.

| Bloque | Guía | Fichero en el que trabajas |
|---|---|---|
| 0 · Antes de empezar | [GUIA-BLOQUE0.pdf](GUIA-BLOQUE0.pdf) | |
| 1 · Almacenamiento con SQLite | [GUIA-BLOQUE1.pdf](GUIA-BLOQUE1.pdf) | `store.go` |
| 2 · API HTTP | [GUIA-BLOQUE2.pdf](GUIA-BLOQUE2.pdf) | `api.go` |
| 3 · Concurrencia y transacciones | [GUIA-BLOQUE3.pdf](GUIA-BLOQUE3.pdf) | `store.go` |

## Para empezar

```bash
git clone https://github.com/GoMalaga/workshops.git
cd workshops/go2
./check-env.sh        # comprueba Go 1.23+ y compila SQLite (un minuto la primera vez)
go run ./cmd/hello    # crea, escribe y lee una base de datos de prueba: si dice ok, listo
```

Lo demás está en la guía del bloque 0.

## Checkpoints

`go2-step0` (punto de partida), `go2-step1` (bloque 1), `go2-step2` (bloques 1
y 2), `go2-final` (todo). `git fetch --tags` y `git checkout <tag>`.
