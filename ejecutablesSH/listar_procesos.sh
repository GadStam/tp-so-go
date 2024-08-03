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

KERNEL_URL="http://$KERNEL_HOST:$KERNEL_PORT/process"

echo "URL: $KERNEL_URL"

curl -X GET "$KERNEL_URL"
