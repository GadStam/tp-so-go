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
if [ "$#" -ne 1 ]; then
    echo "Uso: $0 <PID>"
    exit 1
fi

# Asignar los argumentos a variables
PID="$1"

# URL del servidor
KERNEL_URL="http://$KERNEL_HOST:$KERNEL_PORT/process?pid=$PID"


# Imprimir la URL y el cuerpo JSON para depuración
echo "URL: $KERNEL_URL"

# Realizar la petición DELETE con curl
curl -X DELETE "$KERNEL_URL"
