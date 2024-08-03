#!/bin/bash

if [ -z "$KERNEL_PORT" ]; then
    echo "No se ha definido la variable KERNEL_PORT"
    echo "Usando puerto por defecto 8080"
    KERNEL_PORT=8080
fi

if [ -z "$KERNEL_HOST" ]; then
    echo "No se ha definido la variable KERNEL_HOST"
    echo "Usando HOST por defecto localhost"
    KERNEL_HOST=localhost
fi

KERNEL_URL="http://$KERNEL_HOST:$KERNEL_PORT"

# Verificar si se pasó el argumento necesario
if [ "$#" -ne 1 ]; then
    echo "Uso: $0 <PID>"
    exit 1
fi

# Asignar el argumento a la variable PID
PID="$1"

# Construir la URL completa
URL="$KERNEL_URL/process/$PID"

# Realizar la petición GET con curl
curl -X GET "$URL"
