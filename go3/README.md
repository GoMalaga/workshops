# Taller de Go · Sesión 3: de la consulta a la API

La taquilla de las sesiones 1 y 2, con la base de datos ya conectada. En esta
sesión repasas SQL sobre sus tablas y la pones detrás de una API HTTP. Para
casa queda un ejercicio opcional: un endpoint de cero, desde la consulta hasta
la respuesta.

No hace falta haber terminado la sesión 2: el almacenamiento viene resuelto.

## Las guías

Hay una guía por bloque y una más con el ejercicio para casa. Cada una explica lo que viene hecho y después da el
paso a paso, con el test que comprueba cada paso.

| Bloque | Guía | Fichero en el que trabajas |
|---|---|---|
| 0 · Antes de empezar | [GUIA-BLOQUE0.pdf](GUIA-BLOQUE0.pdf) | |
| 1 · SQL sobre la taquilla | [GUIA-BLOQUE1.pdf](GUIA-BLOQUE1.pdf) | `consultas.go` |
| 2 · API HTTP | [GUIA-BLOQUE2.pdf](GUIA-BLOQUE2.pdf) | `api.go` |
| Ejercicio para casa · Un endpoint de punta a punta | [GUIA-EJERCICIO.pdf](GUIA-EJERCICIO.pdf) | `store.go` y `api.go` |

## Para empezar

```bash
git clone https://github.com/GoMalaga/workshops.git
cd workshops/go3
./check-env.sh        # comprueba Go 1.23+ y compila SQLite (un minuto la primera vez)
go run ./cmd/hello    # crea, escribe y lee una base de datos de prueba: si dice ok, listo
```

Lo demás está en la guía del bloque 0.

## Checkpoints

`go3-step0` (punto de partida), `go3-step1` (bloque 1), `go3-step2` (bloques 1
y 2), `go3-final` (el ejercicio para casa resuelto). `git fetch --tags` y `git checkout <tag>`.
