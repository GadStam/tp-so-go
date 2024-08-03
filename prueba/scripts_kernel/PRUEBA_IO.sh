#!/bin/bash

if [ -z "$KERNEL_PORT" ]; then
    echo "The KERNEL_PORT is not set"
    echo "Using default port 8080"
    KERNEL_PORT=8080
fi

if [ -z "$KERNEL_HOST" ]; then
    echo "The KERNEL_HOST is not set"
    echo "Using default host localhost"
    KERNEL_HOST=localhost
fi

curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 1,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/IO_A"
}'

curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 2,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/IO_B"
}'

curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 3,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/IO_C"
}'