# Distributed System Simulation Project - Kernel Module

## Overview

This project aims to simulate a distributed system, focusing on process scheduling, system request handling, and efficient management of memory and file systems. The Kernel module is a crucial component of this simulation, responsible for managing the execution of various processes generated through its API.

## Kernel Module Responsibilities

1. Process Management
   - Initiate system processes
   - Manage process lifecycle
   - Handle process state transitions

2. Resource Management
   - Manage system resources as defined in the configuration file
   - Handle WAIT and SIGNAL operations for resources

3. I/O Interface Management
   - Manage connections with dynamically connected I/O interfaces
   - Handle I/O requests from processes

4. Memory and CPU Interaction
   - Manage requests to the Memory module for process creation and deletion
   - Schedule process execution on the CPU module

## Key Features

1. Multi-threaded Architecture
   - Allows concurrent handling of requests from different modules

2. Process Control Block (PCB)
   - Manages administrative information for each process
   - Includes execution context for CPU dispatch

3. Long-term Scheduler
   - Handles process creation and deletion requests

4. Short-term Scheduler
   - Implements FIFO, Round Robin, and Virtual Round Robin algorithms
   - Manages process state transitions and CPU dispatching

5. Resource Management
   - Tracks available system resources
   - Manages resource allocation and deallocation

6. I/O Interface Handling
   - Dynamically manages connections with I/O interfaces
   - Handles I/O requests and process blocking/unblocking

## API Operations

- Start process
- End process
- Get process status
- Start scheduling
- Stop scheduling
- List processes

## Implementation Guidelines

1. Follow the iterative incremental development methodology
2. Implement the required logging system
3. Ensure proper handling of inter-module communications (Memory, CPU, I/O interfaces)
4. Implement the 5-state process diagram for scheduling

## Note

This module is estimated to comprise approximately 35% of the overall project workload. Demonstrating theoretical knowledge and practical implementation of this module is crucial for project approval.

# Distributed System Simulation Project - CPU Module

## Overview

The CPU module in this project simulates a simplified version of a real CPU's instruction cycle. It is responsible for interpreting and executing instructions from the Execution Contexts received from the Kernel.

## Key Components

1. Instruction Cycle
2. Memory Management Unit (MMU)
3. Translation Lookaside Buffer (TLB)
4. CPU Registers

## Instruction Cycle

The CPU implements a simplified instruction cycle with the following steps:

1. Fetch: Retrieve the next instruction from Memory using the Program Counter.
2. Decode: Interpret the instruction and determine if address translation is needed.
3. Execute: Perform the operation specified by the instruction.
4. Check Interrupt: Verify if the Kernel has sent an interrupt for the current process.

## Memory Management Unit (MMU)

The MMU is responsible for translating logical addresses to physical addresses. It uses a paging scheme where logical addresses are composed of:

[page_number | offset]

The translation can be performed as follows:
- page_number = floor(logical_address / page_size)
- offset = logical_address - page_number * page_size

## Translation Lookaside Buffer (TLB)

The TLB is implemented to speed up the translation of logical addresses to physical addresses. Its structure includes:

[pid | page | frame]

TLB configuration:
- The number of entries and replacement algorithm are specified in the CPU configuration file.
- Number of entries: Integer (0 disables TLB)
- Replacement algorithms: FIFO or LRU

## CPU Registers

The CPU uses a set of registers to model the operation of a real CPU, including:
- General-purpose registers
- Special-purpose registers (e.g., Program Counter)

## Instruction Set

The CPU supports various instructions, including but not limited to:
- SET, MOV_IN, MOV_OUT: Memory operations
- SUM, SUB: Arithmetic operations
- JNZ: Conditional jump
- RESIZE: Memory allocation
- WAIT, SIGNAL: Resource management
- I/O operations (e.g., IO_STDIN_READ, IO_STDOUT_WRITE)
- File system operations (e.g., IO_FS_CREATE, IO_FS_DELETE)
- EXIT: Process termination

## Interaction with Other Modules

- Receives Execution Contexts from the Kernel
- Fetches instructions from Memory
- Handles interrupts from the Kernel
- Returns updated Execution Contexts to the Kernel upon process completion, blocking, or preemption

## Implementation Notes

1. The CPU module comprises approximately 15% of the overall project workload.
2. Implement proper error handling for invalid instructions or memory access.
3. Ensure accurate updating of the Execution Context throughout the instruction cycle.
4. Implement the TLB with the specified replacement algorithms.
5. Handle logical to physical address translation correctly, considering the paging scheme.

# Distributed System Simulation Project - Memory Module

## Overview

The Memory module is responsible for managing the system's memory, including instruction storage, user space, and page tables. It implements a simple paging scheme and handles various memory-related operations.

## Key Components

1. Instruction Memory
2. User Space Memory
3. Page Tables
4. Communication interfaces with Kernel, CPU, and I/O Interfaces

## Instruction Memory

- Stores instructions from pseudo-code files
- Provides instructions one at a time to the CPU upon request
- Simulates memory access delay as specified in the configuration file

## Memory Scheme

### Structures

1. User Space: Represented by a contiguous array of bytes
2. Page Tables: Manage the mapping between logical and physical addresses

Important: The entire memory space dedicated to user space must be contiguous. Any implementation that doesn't follow this rule will result in immediate project failure.

Note: The total memory size will always be a multiple of the page size.

## Communication Interfaces

### Process Creation (Kernel only)

- Creates necessary administrative structures for a new process

### Process Termination (Kernel only)

- Frees memory space occupied by the terminated process
- Marks frames as free without overwriting their content

### Page Table Access

- Responds with the frame number corresponding to the queried page

### Process Size Adjustment

Handles two scenarios:
1. Process Expansion
   - Expands the process size at the end
   - Responds with "Out Of Memory" error if unable to allocate required frames
2. Process Reduction
   - Reduces the process size from the end
   - Frees unused pages as necessary

### User Space Access (CPU and I/O Interfaces)

- Handles read and write requests to physical addresses
- Supports requests that may span multiple pages
- Simulates access delay as specified in the configuration file

## Implementation Notes

1. The Memory module comprises approximately 20% of the overall project workload.
2. Ensure proper implementation of the paging scheme.
3. Handle multi-page operations correctly.
4. Implement accurate simulation of memory access delays.
5. Maintain proper synchronization for concurrent access from different modules.


# Distributed System Simulation Project - I/O Interface Module

## Overview

The I/O Interface module simulates various input/output devices such as keyboards, mice, disks, monitors, and printers. It handles operations requested by the Kernel for specific processes, processing them one at a time in the order of arrival.

## Key Components

1. Generic Interfaces
2. STDIN Interfaces
3. STDOUT Interfaces
4. DialFS Interfaces

## Common Configuration

Each I/O Interface requires two parameters at startup:
1. Name: A unique identifier for the interface within the system
2. Configuration File: Contains specific settings for the interface

## Interface Types

### 1. Generic Interfaces

- Simplest type of interface
- Waits for a specified number of work units upon request
- Accepts instruction: IO_GEN_SLEEP
- Configuration properties: type, unit_work_time, ip_kernel, port_kernel

### 2. STDIN Interfaces

- Waits for user input via keyboard
- Stores input in memory at the specified physical address
- Accepts instruction: IO_STDIN_READ
- Configuration properties: type, ip_kernel, port_kernel, ip_memory, port_memory

### 3. STDOUT Interfaces

- Reads from a physical memory address and displays the result
- Always consumes one unit of unit_work_time
- Accepts instruction: IO_STDOUT_WRITE
- Configuration properties: type, unit_work_time, ip_kernel, port_kernel, ip_memory, port_memory

### 4. DialFS Interfaces

- Most complex interface type
- Interacts with a file system (DialFS) implemented by the project group
- Always consumes one unit of unit_work_time
- Accepts instructions: IO_FS_CREATE, IO_FS_DELETE, IO_FS_TRUNCATE, IO_FS_WRITE, IO_FS_READ
- Configuration properties: type, unit_work_time, ip_kernel, port_kernel, ip_memory, port_memory, dialfs_path, dialfs_block_size, dialfs_block_count

## DialFS File System

DialFS is a simple implementation of a Contiguous Allocation File System, simulated using the following files:

1. blocks.dat: Represents the file system blocks
2. bitmap.dat: Bitmap indicating free and occupied blocks
3. Metadata files: JSON files for each file in the FS, containing initial block and size information

### Key Features

- File Creation: Initially occupies one block, even for empty files
- Compaction: Regroups files to create contiguous free space when needed
- Truncation: Allows file size adjustment using IO_FS_TRUNCATE

## Implementation Notes

1. Ensure proper handling of different interface types and their specific instructions
2. Implement accurate simulation of work unit delays
3. For DialFS, ensure correct implementation of the bitmap-based block allocation
4. Handle file system compaction when necessary
5. Maintain proper synchronization for concurrent access from different processes

# RESUME
For a comprehensive understanding of this project, please refer to the detailed [project documentation](https://docs.google.com/document/d/136ehPcR28FeS1RzuW58t2IceYQ2oLvYzpqWKLpUJo10/edit?usp=drive_link).

## Overview

This project aims to simulate a distributed system, focusing on process scheduling, system request handling, and efficient management of memory and file systems. The system is composed of several modules working together to create a simplified but functional distributed computing environment.

## Modules

### 1. Kernel Module

#### Responsibilities
- Process Management
- Resource Management
- I/O Interface Management
- Memory and CPU Interaction

#### Key Features
- Multi-threaded Architecture
- Process Control Block (PCB) Management
- Long-term and Short-term Schedulers
- Resource and I/O Interface Handling

#### API Operations
- Start process
- End process
- Get process status
- Start/Stop scheduling
- List processes

### 2. CPU Module

#### Key Components
- Instruction Cycle (Fetch, Decode, Execute, Check Interrupt)
- Memory Management Unit (MMU)
- Translation Lookaside Buffer (TLB)
- CPU Registers

#### Instruction Set
Includes operations for memory access, arithmetic, control flow, resource management, and I/O operations.

### 3. Memory Module

#### Key Components
- Instruction Memory
- User Space Memory (contiguous array of bytes)
- Page Tables

#### Operations
- Process Creation and Termination
- Page Table Access
- Process Size Adjustment
- User Space Access

### 4. I/O Interface Module

#### Interface Types
1. Generic Interfaces
2. STDIN Interfaces
3. STDOUT Interfaces
4. DialFS Interfaces

#### DialFS File System
A simple Contiguous Allocation File System implementation.

***

## Deployment Instructions

### Prerequisites
- Linux environment (can be run in a virtual machine)
- Git
- Go programming language

### Setup

1. Set up the Linux environment:
   - If using a VM, ensure it's properly configured

2. Install Go:
  ```bash
  git clone https://github.com/sisoputnfrba/entorno-vms --depth=1
  cd entorno-vms
  sudo bash -x base-server.sh
  ./golang.sh
  ```

3. Clone the project repository:
  ```bash
 git clone https://github.com/GadStam/tp-so-go.git
  ```

4. Configure IP addresses:
   
   ```bash
   cd /path/to/tp-so-go/ejecutablesSH
   go run script_ip.go -es "ipIO" -cpu "ipCPU" -kernel "ipKERNEL" -memoria "ipMEMORIA"
    ```
Note: If running locally, use "localhost" for all modules.

### Running the Modules

1. KERNEL
 ```bash
   go build -o kernel.exe kernel.go
./kernel.exe /path/to/tp-so-go/kernel/KERNELconfigs/DESIRED_CONFIG.json
  ```
2. CPU
 ```bash
  go build -o cpu.exe cpu.go
./cpu.exe /path/to/tp-so-go/cpu/CPUconfigs/DESIRED_CONFIG.json
  ```
3. MEMORY
 ```bash
  go build -o memoria.exe memoria.go
./memoria.exe /path/to/tp-so-go/memoria/MEMconfigs/DESIRED_CONFIG.json
  ```
4. I/O (run as needed)
  ```bash
    go build -o entradasalida.exe entradasalida.go
./entradasalida.exe INTERFACE_NAME /path/to/tp-so-go/entradasalida/ioConfigs/INTERFACE_NAME.json
  ```


### Running Tests

For a complete list of available tests and their descriptions, please refer to the [test documentation](https://docs.google.com/document/d/1845nrvfTM9Juw4MVEQp6MYg3hYevmSXJAQjKn1Bdevc/edit?usp=drive_link).


1. Navigate to the test scripts folder:
  ```bash
  cd /path/to/tp-so-go/prueba/scripts_kernel
  ```
2. Execute the desired test script:
  ```bash
  ./TEST_NAME.sh
  ```
Example: `./PRUEBA_PLANI.sh`



## Implementation Notes

- The Kernel module comprises approximately 35% of the overall project workload.
- The CPU module comprises approximately 15% of the overall project workload.
- The Memory module comprises approximately 20% of the overall project workload.
- Ensure proper synchronization for concurrent access in all modules.
- Implement accurate simulation of delays and work units as specified in configuration files.
- For the DialFS, ensure correct implementation of the bitmap-based block allocation and handle file system compaction when necessary.

## Additional Notes

- Ensure all paths are correctly set according to your system's directory structure.
- Make sure all necessary permissions are granted for executing the scripts and binaries.
- It's recommended to run the modules in separate terminal windows for easier monitoring and debugging.


## Contributors

This project wouldn't be possible without the valuable contributions of the following individuals:

* **Alan Garber** ([@AlanGarber](https://github.com/AlanGarber))
* **Fausto Oliva** ([@FaustoOliva](https://github.com/FaustoOliva))
* **Gonzalo Vaserman** ([@gonzivaser](https://github.com/gonzivaser))
* **Franco Ysraelit** ([@FrancoYsrraelit](https://github.com/FrancoYsrraelit))

Thank you for your dedication and hard work :)
