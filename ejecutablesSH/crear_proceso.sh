#!/bin/bash

if [ -z "$KERNEL_PORT" ]; then
    echo "No se ha definido la variable KERNEL_PORT"
    echo "Usando puerto por defecto 8080"
    KERNEL_PORT=8080
fi

# Verificar si se ha definido la variable KERNEL_HOST
if [ -z "$KERNEL_HOST" ]; then
    echo "No se ha definido la variable KERNEL_HOST"
    echo "Usando HOST por defecto localhost"
    KERNEL_HOST=localhost
fi


# Verificar si se pasaron los argumentos necesarios
if [ "$#" -ne 2 ]; then
    echo "Uso: $0 <PID> <PATH>"
    exit 1
fi

# Asignar los argumentos a variables
PID="$1"
FILE_PATH="$2"

# URL del servidor
KERNEL_URL="http://$KERNEL_HOST:$KERNEL_PORT/process"

# Cuerpo JSON
BODY="{\"pid\": $PID, \"path\": \"$FILE_PATH\"}"

# Imprimir la URL y el cuerpo JSON para depuración
echo "URL: $KERNEL_URL"
echo "Cuerpo JSON: $BODY"

# Realizar la petición PUT con curl
curl -X PUT "$KERNEL_URL" -H "Content-Type: application/json" -d "$BODY"
