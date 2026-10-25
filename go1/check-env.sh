#!/usr/bin/env bash
# Bloque 0: valida el entorno antes de empezar.
set -e

if ! command -v go >/dev/null 2>&1; then
  echo "❌ Go no está instalado. Instálalo desde https://go.dev/dl/"
  exit 1
fi

VERSION="$(go version | awk '{print $3}' | sed 's/go//')"
MAJOR="$(echo "$VERSION" | cut -d. -f1)"
MINOR="$(echo "$VERSION" | cut -d. -f2)"

if [ "$MAJOR" -lt 1 ] || { [ "$MAJOR" -eq 1 ] && [ "$MINOR" -lt 22 ]; }; then
  echo "❌ Tu versión de Go es $VERSION. Necesitas 1.22 o superior."
  exit 1
fi

echo "✅ Entorno listo. Go $VERSION detectado."
echo "   Ahora ejecuta:  go run ./cmd/hello"
