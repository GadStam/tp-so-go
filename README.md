## Checkpoint

Para cada checkpoint de control obligatorio, se debe crear un tag en el
repositorio con el siguiente formato:

```
checkpoint-{número}
```

Donde `{número}` es el número del checkpoint.

Para crear un tag y subirlo al repositorio, podemos utilizar los siguientes
comandos:

```bash
git tag -a checkpoint-{número} -m "Checkpoint {número}"
git push origin checkpoint-{número}
```

Asegúrense de que el código compila y cumple con los requisitos del checkpoint
antes de subir el tag.

### Checkpoint 1

- [x] Familiarizarse con Linux y su consola, el entorno de desarrollo y el repositorio.
- [x] Aprender a utilizar las Commons, principalmente las funciones para listas, archivos de configuración y logs.
- [x] Definir el Protocolo de Comunicación.
- [x] Todas las API del módulo Kernel definidas por la cátedra están creadas y retornan datos hardcodeados.
- [x] Todos los módulos están creados y son capaces de inicializarse con al menos una API.

### Checkpoint 2

- Módulo Kernel:
  
    - [x] Es capaz de crear un PCB y planificarlo por FIFO y RR.
    - [x] Es capaz de enviar un proceso a la CPU para que sea procesado.

- Módulo CPU:
  
    - [x] Se conecta a Kernel y recibe un PCB.
    - [x] Es capaz de conectarse a la memoria y solicitar las instrucciones.
    - [x] Es capaz de ejecutar un ciclo básico de instrucción.
    - [x] Es capaz de resolver las operaciones: SET, SUM, SUB, JNZ e IO_GEN_SLEEP.

- Módulo Memoria:
  
    - [x] Se encuentra creado y acepta las conexiones.
    - [x] Es capaz de abrir los archivos de pseudocódigo y envía las instrucciones al CPU.

- Módulo Interfaz I/O:
- 
    - [x] Se encuentra desarrollada la Interfaz Genérica.


### Checkpoint 3

- Módulo Kernel:

    - [x] Es capaz de planificar por VRR.
    - [x] Es capaz de realizar manejo de recursos.
    - [x] Es capaz de manejar el planificador de largo plazo


- Módulo CPU:

    - [ ] Es capaz de resolver las operaciones: MOV_IN, MOV_OUT, RESIZE, COPY_STRING, IO_STDIN_READ, IO_STDOUT_WRITE.

- Módulo Memoria:

    - [x] Se encuentra completamente desarrollada.

- Módulo Interfaz I/O:

    - [x] Se encuentran desarrolladas las interfaces STDIN y STDOUT.

### Logs mínimos y obligatorios

- Módulo Kernel:

    - [x] Creación de Proceso
    - [x] Fin de Proceso
    - [x] Cambio de Estado
    - [x] Motivo de Bloqueo
    - [x] Fin de Quantum
    - [x] Ingreso a Ready


- Módulo CPU:

    - [x] Fetch Instrucción
    - [x] Instrucción Ejecutada

    - [ ] TLB Hit
    - [ ] TLB Miss

    - [x] Obtener Marco
    - [x] Lectura/Escritura Memoria

- Módulo Memoria:


    - [x] Se Creación / destrucción de Tabla de Páginas
    - [x] Acceso a Tabla de Páginas
    - [x] Ampliación de Proceso
    - [x] Reducción de Proceso
    - [x] Acceso a espacio de usuario


- Módulo Interfaz I/O:

    - [X] Todos - Operación
    - [ ] DialFS - Crear Archivo
    - [ ] DialFS - Eliminar Archivo
    - [ ] DialFS - Truncar Archivo
    - [ ] DialFS - Leer Archivo
    - [ ] DialFS - Escribir Archivo


## Entregas finales

- [ ] Finalizar el desarrollo de todos los procesos.
- [ ] Probar de manera intensiva el TP en un entorno distribuido.
- [ ] Todos los componentes del TP ejecutan los requerimientos de forma integral.

