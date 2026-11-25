#!/usr/bin/env bash
# Bloque 0: valida el entorno y compila SQLite una vez (tarda un minuto).
set -e

if ! command -v go >/dev/null 2>&1; then
  echo "❌ Go no está instalado. Instálalo desde https://go.dev/dl/"
  exit 1
fi

VERSION="$(go version | awk '{print $3}' | sed 's/go//')"
MAJOR="$(echo "$VERSION" | cut -d. -f1)"
MINOR="$(echo "$VERSION" | cut -d. -f2)"

if [ "$MAJOR" -lt 1 ] || { [ "$MAJOR" -eq 1 ] && [ "$MINOR" -lt 23 ]; }; then
  echo "❌ Tu versión de Go es $VERSION. Necesitas 1.23 o superior."
  exit 1
fi
echo "✅ Go $VERSION detectado."

echo "⏳ Descargando y compilando el driver de SQLite (solo la primera vez)..."
go build ./...
echo "✅ Entorno listo."
echo "   Ahora ejecuta:  go run ./cmd/hello"
