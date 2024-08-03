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
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/TRIANGLE"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 2,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/SQUARE"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 3,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/CIRCLE"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 4,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/PIRAMID"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 5,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/CUBE"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 6,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/SPHERE"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 7,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/TRIANGLE"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 8,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/SQUARE"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 9,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/CIRCLE"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 10,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/PIRAMID"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 11,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/CUBE"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 12,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/SPHERE"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 13,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/TRIANGLE"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 14,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/SQUARE"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 15,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/CIRCLE"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 16,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/PIRAMID"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 17,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/CUBE"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 18,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/SPHERE"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 19,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/TRIANGLE"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 20,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/SQUARE"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 21,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/CIRCLE"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 22,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/PIRAMID"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 23,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/CUBE"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 24,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/SPHERE"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 25,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/TRIANGLE"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 26,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/SQUARE"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 27,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/CIRCLE"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 28,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/PIRAMID"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 29,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/CUBE"
}' 
curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
--header 'Content-Type: application/json' \
--data '{
    "pid": 30,
    "path": "/home/utnso/tp-2024-1c-Panza_confianza/prueba/scripts_memoria/SPHERE"
}'