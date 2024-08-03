package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"

	"path/filepath"

	"strings"
	"time"

	"github.com/sisoputnfrba/tp-golang/entradasalida/globals"
)

func ConfigurarLogger(interfazNombre string, config *globals.Config) {
	var logFile *os.File
	var err error

	if config.Tipo == "DialFS" {
		// Para DialFS, crear un archivo de log específico
		logFileName := fmt.Sprintf("%s_DialFS.log", interfazNombre)
		logFile, err = os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		if err != nil {
			panic(err)
		}
	} else {
		// Para otros tipos, usar el archivo de log general
		logFile, err = os.OpenFile("entradasalida.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		if err != nil {
			panic(err)
		}
	}

	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)

	// Configurar el prefijo del log para incluir el nombre de la interfaz
	log.SetPrefix(fmt.Sprintf("[%s] ", interfazNombre))
	log.SetFlags(log.Ldate | log.Ltime)
}

func IniciarConfiguracion(filePath string) *globals.Config {
	var config *globals.Config
	configFile, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	return config
}

// datos unicos de cada proceso
type ProcessData struct {
	Pid             int
	LengthREG       int
	DireccionFisica []int
}

var processDataMap sync.Map

/*---------------------------------------------------- STRUCTS ------------------------------------------------------*/
type BodyRequestPort struct {
	Nombre string `json:"nombre"`
	Port   int    `json:"port"`
	Type   string `json:"type"`
}

type BodyRequestRegister struct {
	Length  int   `json:"lengthREG"`
	Address []int `json:"dirFisica"`
	Pid     int   `json:"iopid"`
}

type BodyRequestInput struct {
	Pid     int    `json:"pid"`
	Input   string `json:"input"`
	Address []int  `json:"address"` //Esto viene desde kernel
}

type BodyContent struct {
	Content string `json:"content"`
}

type InterfazIO struct {
	Nombre string         // Nombre único
	Config globals.Config // Configuración
}

type Payload struct {
	IO  int
	Pid int
}

type MemoryRequest struct {
	PID     int    `json:"pid"`
	Address []int  `json:"address"`
	Size    int    `json:"size,omitempty"` //Si es 0, se omite (Util para creacion y terminacion de procesos)
	Data    []byte `json:"data,omitempty"` //Si es 0, se omite Util para creacion y terminacion de procesos)
	Type    string `json:"type"`           //Si es 0, se omite Util para creacion y terminacion de procesos)
	Port    int    `json:"port,omitempty"`
}

type FSstructure struct {
	FileName      string `json:"filename"`
	FSInstruction string `json:"fsinstruction"`
	FSRegTam      int    `json:"fsregtam"`
	FSRegDirec    []int  `json:"fsregdirec"`
	FSRegPuntero  int    `json:"fsregpuntero"`
}

type FileContent struct {
	InitialBlock int `json:"initial_block"`
	Size         int `json:"size"`
	FileName     string
}

type Bitmap struct {
	bits       []int
	blockCount int
	blockSize  int
}

type BlockFile struct {
	FilePath    string
	BlocksSize  int
	BlocksCount int
	FreeBlocks  []bool // Un slice para rastrear si un bloque está libre
}

/*--------------------------- ESTRUCTURA DEL METADATA -----------------------------*/

var metaDataStructure []FileContent

/*--------------------------- NOMBRE DEL ARCHIVO E INSTRUCCION -----------------------------*/

var fileName string
var fsInstruction string
var fsRegTam int
var fsRegDirec []int
var fsRegPuntero int

/*--------------------------------------------------- VAR GLOBALES ------------------------------------------------------*/

var GLOBALmemoryContent string
var config *globals.Config

/*-------------------------------------------- INICIAR CONFIGURACION ------------------------------------------------------*/

func LoadConfig(filename string) (*globals.Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config globals.Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func Iniciar(w http.ResponseWriter, r *http.Request) {
	var payload Payload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Error al decodificar los datos JSON", http.StatusInternalServerError)
		return
	}

	N := payload.IO
	pidExecutionProcess := payload.Pid

	interfaceName := os.Args[1]

	processData, _ := getProcessData(payload.Pid)

	pathToConfig := os.Args[2]

	config, err = LoadConfig(pathToConfig)
	if err != nil {
		log.Fatalf("Error al cargar la configuración desde '%s': %v", pathToConfig, err)
	}

	Interfaz := &InterfazIO{
		Nombre: interfaceName,
		Config: *config,
	}

	switch Interfaz.Config.Tipo {
	case "GENERICA":
		log.Printf("PID: %d - Operacion: IO_GEN_SLEEP", pidExecutionProcess)
		duracion := Interfaz.IO_GEN_SLEEP(N)
		time.Sleep(duracion)

	case "STDIN":
		log.Printf("PID: %d - Operacion: IO_STDIN_READ", pidExecutionProcess)
		Interfaz.IO_STDIN_READ(processData.DireccionFisica, processData.LengthREG, pidExecutionProcess)

	case "STDOUT":
		log.Printf("PID: %d - Operacion: IO_STDOUT_WRITE", pidExecutionProcess)
		Interfaz.IO_STDOUT_WRITE(processData.DireccionFisica, processData.LengthREG, pidExecutionProcess)

	case "DialFS":
		createDirectory(Interfaz.Config.PathDialFS)
		Interfaz.FILE_SYSTEM(pidExecutionProcess)

	default:
		log.Fatalf("Tipo de interfaz desconocido: %s", Interfaz.Config.Tipo)
	}
}

func createDirectory(path string) {
	pahDialFS := path + "/FS"
	err := os.MkdirAll(pahDialFS, 0755)
	if err != nil {
		fmt.Printf("Error al crear la carpeta: %v\n", err)
		return
	}
}

/*-------------------------------------------------- ENDPOINTS ------------------------------------------------------*/

func FinalizarProceso(pid int) {
	kernelURL := fmt.Sprintf("http://%s:%d/process?pid=%d", config.IPKernel, config.PuertoKernel, pid)
	req, err := http.NewRequest("DELETE", kernelURL, nil)
	if err != nil {
		log.Fatalf("Error al crear la solicitud: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error al enviar la solicitud al módulo de kernel: %v", err)
	}
	defer resp.Body.Close()
}

func SendPortOfInterfaceToKernel(nombreInterfaz string, config *globals.Config) error {
	kernelURL := fmt.Sprintf("http://%s:%d/SendPortOfInterfaceToKernel", config.IPKernel, config.PuertoKernel)

	interfaceData := BodyRequestPort{
		Nombre: nombreInterfaz,
		Port:   config.Puerto,
		Type:   config.Tipo,
	}

	interfaceDataJSON, err := json.Marshal(interfaceData)
	if err != nil {
		log.Fatalf("Error al codificar el puerto a JSON: %v", err)
	}

	resp, err := http.Post(kernelURL, "application/json", bytes.NewBuffer(interfaceDataJSON))
	if err != nil {
		return fmt.Errorf("error al enviar la solicitud al módulo kernel: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error en la respuesta del módulo kernel: %v", resp.StatusCode)
	}

	return nil
}

// STDOUT, FS_WRITE  leer en memoria y traer lo leido con "ReceiveContentFromMemory"
func SendAdressToMemory(ReadRequest MemoryRequest) error {
	memoriaURL := fmt.Sprintf("http://%s:%d/readMemory", config.IPMemoria, config.PuertoMemoria)

	adressResponseTest, err := json.Marshal(ReadRequest)
	if err != nil {
		log.Fatalf("Error al serializar el address: %v", err)
	}

	resp, err := http.Post(memoriaURL, "application/json", bytes.NewBuffer(adressResponseTest))
	if err != nil {
		log.Fatalf("Error al enviar la solicitud al módulo de memoria: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Error en la respuesta del módulo de memoria: %v", resp.StatusCode)
	}

	return nil
}

// STDIN, FS_READ  escribir en memoria
func SendInputToMemory(pid int, input string, address []int) error {
	bodyRequest := MemoryRequest{
		PID:     pid,
		Data:    []byte(input),
		Address: address,
	}

	memoriaURL := fmt.Sprintf("http://%s:%d/writeMemory", config.IPMemoria, config.PuertoMemoria)

	inputResponseTest, err := json.Marshal(bodyRequest)
	if err != nil {
		log.Fatalf("Error al serializar el Input: %v", err)
	}

	resp, err := http.Post(memoriaURL, "application/json", bytes.NewBuffer(inputResponseTest))
	if err != nil {
		//fmt.Printf("Error al enviar la solicitud al módulo de memoria: %v", err)
		FinalizarProceso(pid)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Error en la respuesta del módulo de memoria: %v", resp)
	}

	return nil
}

func RecieveREG(w http.ResponseWriter, r *http.Request) {
	var requestRegister BodyRequestRegister

	err := json.NewDecoder(r.Body).Decode(&requestRegister)
	if err != nil {
		http.Error(w, "Error decoding JSON data", http.StatusInternalServerError)
		return
	}

	processData := ProcessData{
		Pid:             requestRegister.Pid,
		LengthREG:       requestRegister.Length,
		DireccionFisica: requestRegister.Address,
	}
	processDataMap.Store(requestRegister.Pid, processData)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("length received: %d", requestRegister.Length)))
}

func RecieveFSDataFromKernel(w http.ResponseWriter, r *http.Request) {
	var fsStructure FSstructure

	err := json.NewDecoder(r.Body).Decode(&fsStructure)
	if err != nil {
		http.Error(w, "Error decoding JSON data", http.StatusInternalServerError)
		return
	}

	fileName = fsStructure.FileName
	fsInstruction = fsStructure.FSInstruction
	fsRegTam = fsStructure.FSRegTam
	fsRegDirec = fsStructure.FSRegDirec
	fsRegPuntero = fsStructure.FSRegPuntero

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Content received correctly"))
}

func ReceiveContentFromMemory(w http.ResponseWriter, r *http.Request) {
	var content BodyContent
	err := json.NewDecoder(r.Body).Decode(&content)

	if err != nil {
		http.Error(w, "Error decoding JSON data", http.StatusInternalServerError)
		return
	}

	GLOBALmemoryContent = content.Content

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Content received correctly"))
}

/*----------------------------------------- FUNCIONES AUXILIARES --------------------------------------------------*/

func getProcessData(pid int) (ProcessData, bool) {
	data, ok := processDataMap.Load(pid)
	if !ok {
		return ProcessData{}, false
	}
	return data.(ProcessData), true
}

/*---------------------------------------------- INTERFACES -------------------------------------------------------*/

// INTERFAZ STDOUT (IO_STDOUT_WRITE)
func (Interfaz *InterfazIO) IO_STDOUT_WRITE(address []int, length int, pid int) {

	//var Bodyadress BodyAdress
	req := MemoryRequest{
		PID:     pid,
		Address: address,
		Size:    length,
		Type:    "IO",
		Port:    config.Puerto,
	}

	err1 := SendAdressToMemory(req)
	if err1 != nil {
		log.Fatalf("Error al leer desde la memoria: %v", err1)
	}

	time.Sleep(time.Duration(config.UnidadDeTiempo) * time.Millisecond)
	fmt.Println(GLOBALmemoryContent)
}

// INTERFAZ STDIN (IO_STDIN_READ)
func (Interfaz *InterfazIO) IO_STDIN_READ(address []int, lengthREG int, pid int) {
	//var BodyInput BodyRequestInput
	var input string

	//var inputMenorARegLongitud string

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Ingrese por teclado: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Error al leer desde stdin: %v", err)
	}
	input = strings.TrimSpace(input)

	if len(input) > lengthREG {
		input = input[:lengthREG]
		//log.Println("El texto ingresado es mayor al tamaño del registro, se truncará a: ", input)
	} else if len(input) < lengthREG {
		fmt.Print("El texto ingresado es menor, porfavor ingrese devuelta: ")

		complemento, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Error al leer desde stdin: %v", err)
		}
		complemento = strings.TrimSpace(complemento)
		input += complemento
		if len(input) > lengthREG {
			input = input[:lengthREG]
		}

	}

	// Guardar el texto en la memoria en la dirección especificada
	err1 := SendInputToMemory(pid, input, address)
	if err1 != nil {
		log.Printf("Error al escribir en la memoria: %v", err)
	}
}

// INTERFAZ GENERICA (IO_GEN_SLEEP)
func (interfaz *InterfazIO) IO_GEN_SLEEP(n int) time.Duration {
	return time.Duration(n*interfaz.Config.UnidadDeTiempo) * time.Millisecond
}

// INTERFAZ FILE SYSTEM
func (interfaz *InterfazIO) FILE_SYSTEM(pid int) {

	pathDialFS := interfaz.Config.PathDialFS + "/FS"
	blocksSize := interfaz.Config.TamanioBloqueDialFS
	blocksCount := interfaz.Config.CantidadBloquesDialFS
	sizeFile := blocksSize * blocksCount
	bitmapSize := blocksCount / 8
	unitWorkTimeFS := interfaz.Config.UnidadDeTiempo

	// CHEQUEO EXISTENCIA DE ARCHIVOS BLOQUES.DAT Y BITMAP.DAT, DE NO SER ASI, LOS CREO
	ensureExistingMetaDataFiles(pathDialFS)
	EnsureIfFileExists(pathDialFS, blocksSize, blocksCount, sizeFile, bitmapSize)

	switch fsInstruction {
	case "IO_FS_CREATE":
		log.Printf("PID: %d - Crear Archivo: %s", pid, fileName)
		IO_FS_CREATE(pathDialFS, fileName)

	case "IO_FS_DELETE":
		log.Printf("PID: %d - Eliminar Archivo: %s", pid, fileName)
		IO_FS_DELETE(pathDialFS, fileName)

	case "IO_FS_WRITE":
		log.Printf("PID: %d - Operacion: IO_FS_WRITE - Escribir Archivo: %s - Tamaño a Escribir: %d - Puntero Archivo: %d", pid, fileName, fsRegTam, fsRegPuntero)
		IO_FS_WRITE(pathDialFS, fileName, fsRegDirec, fsRegTam, fsRegPuntero, pid)

	case "IO_FS_TRUNCATE":
		log.Printf("PID: %d - Operacion: IO_FS_TRUNCATE", pid)
		IO_FS_TRUNCATE(pathDialFS, fileName, fsRegTam)

	case "IO_FS_READ":
		log.Printf("PID: %d - Operacion: IO_FS_READ - Leer Archivo: %s - Tamaño a Leer: %d - Puntero Archivo: %d", pid, fileName, fsRegTam, fsRegPuntero)
		IO_FS_READ(pathDialFS, fileName, fsRegDirec, fsRegTam, fsRegPuntero, pid)
	}

	time.Sleep(time.Duration(unitWorkTimeFS) * time.Millisecond)
}

/* -------------------------------------------- FUNCIONES DE FS_CREATE ------------------------------------------------------ */

func IO_FS_CREATE(pathDialFS string, fileName string) {
	// CREO ARCHIVO
	crearArchivo(pathDialFS, fileName)

	// ABRO BITMAP Y LO PASO A SLICE
	bitmapFilePath := pathDialFS + "/bitmap.dat"
	bitmap := readAndCopyBitMap(bitmapFilePath)

	// CALCULO PRIMER BIT LIBRE
	firstFreeBlock := firstBitFree(bitmap)

	// SI MI FIRSTFREEBLOCK ES -1, NO HAY BLOQUES LIBRES DISPONIBLES
	if firstFreeBlock == -1 {
		log.Printf("No hay bloques libres disponibles")
	} else {

		// SETEO EL PRIMER BIT LIBRE EN 1
		bitmap.Set(firstFreeBlock)

		showBitmap(bitmap)

		updateBitMap(bitmap, bitmapFilePath)

		// ACTUALIZO EL TAMAÑO A 0 Y EL METADATA
		fileSize := 0
		updateMetaDataFile(pathDialFS, fileName, firstFreeBlock, fileSize)
	}
}

func firstBitFree(bitmap *Bitmap) int {
	for i := 0; i < config.CantidadBloquesDialFS; i++ {
		isFree := !bitmap.Get(i)
		if isFree {
			return i
		}

	}
	return -1
}

func crearArchivo(path string, fileName string) {
	filePath := path + "/" + fileName
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Error al crear el archivo '%s': %v", path, err)
	}
	defer file.Close()
}

/* ------------------------------------------- FUNCIONES DE FS_DELETE ------------------------------------------------------ */

func IO_FS_DELETE(pathDialFS string, fileName string) {
	// PRIMERO ELIMINO EL ARCHIVO
	eliminarArchivo(fileName, pathDialFS)

	// UNA VEZ REMOVIDO EL ARCHIVO, TENGO QUE ACTUALIZAR BITMAP Y ARCHIVO DE BLOQUES
	var fileData FileContent
	fileData, err := dataFileInMetaDataStructure(fileName)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	var blocksToDelete int
	if fileData.Size > config.TamanioBloqueDialFS {
		blocksToDelete = (fileData.Size + config.TamanioBloqueDialFS - 1) / config.TamanioBloqueDialFS
	} else {
		blocksToDelete = 1
	}

	// ABRO BITMAP Y LO PONGO EN SLICE
	bitmapFilePath := pathDialFS + "/bitmap.dat"
	bitmap := readAndCopyBitMap(bitmapFilePath)

	for i := fileData.InitialBlock; i < fileData.InitialBlock+blocksToDelete; i++ {
		bitmap.Remove(i)
	}

	showBitmap(bitmap)

	updateBitMap(bitmap, bitmapFilePath)

	//deleteInBlockFile(pathDialFS, blocksToDelete, fileData.InitialBlock)

	deleteInMetaDataStructure(fileName)
}

func deleteInBlockFile(pathDialFS string, blocksToDelete int, initialBlock int) error {
	// Abre el archivo bloques.dat en modo lectura/escritura
	blocksFilePath := filepath.Join(pathDialFS, "bloques.dat")
	blocksFile, err := os.OpenFile(blocksFilePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo de bloques '%s': %v", blocksFilePath, err)
	}
	defer blocksFile.Close()

	// Crear un buffer de ceros del tamaño de un bloque
	zeroBuffer := make([]byte, config.TamanioBloqueDialFS)

	// Iterar sobre los bloques a "eliminar" (limpiar)
	for i := 0; i < blocksToDelete; i++ {
		// Calcular la posición del bloque
		blockPos := int64(initialBlock+i) * int64(config.TamanioBloqueDialFS)

		// Escribir ceros en la posición del bloque
		_, err := blocksFile.WriteAt(zeroBuffer, blockPos)
		if err != nil {
			return fmt.Errorf("error al limpiar el bloque %d: %v", initialBlock+i, err)
		}
	}

	return nil
}

func eliminarArchivo(fileName string, pathDialFS string) {
	filePath := pathDialFS + "/" + fileName
	err := os.Remove(filePath)
	if err != nil {
		log.Fatalf("Error al eliminar el archivo '%s': %v", pathDialFS, err)
	}
}

/* ----------------------------------------- FUNCIONES DE FS_TRUNCATE ------------------------------------------------------ */

func IO_FS_TRUNCATE(pathDialFS string, fileName string, length int) {
	// VERIFICO EXISTENCIA DE ARCHIVO
	verificarExistenciaDeArchivo(pathDialFS, fileName)
	/*for i, fileContent := range metaDataStructure {
		log.Printf("Archivo: %d ", i)
		log.Printf("Archivo: %s ", fileContent.FileName)
		blocksPerFile := getBlocksFile(fileContent.FileName)
		log.Printf("Bloques del archivo: %d", blocksPerFile)
		log.Printf("Posicion Bloque inicial: %d", fileContent.InitialBlock)
	}*/

	// SACO LA CANTIDAD DE BLOQUES NECESARIOS
	var fileData FileContent
	fileData, err := dataFileInMetaDataStructure(fileName)
	if err != nil {
		// Handle the error, e.g., log it or return it
		log.Printf("Error: %v", err)
		return
	}
	bitmapFilePath := pathDialFS + "/bitmap.dat"
	blocksFilePath := pathDialFS + "/bloques.dat"
	bitmap := readAndCopyBitMap(bitmapFilePath)
	cantBloques := (length + config.TamanioBloqueDialFS - 1) / config.TamanioBloqueDialFS
	//log.Printf("Cantidad de bloques necesarios: %d para el archivo %s", cantBloques, fileName)
	totalFreeBlocks := getTotalFreeBlocks(bitmap)

	if length > fileData.Size {
		areFree := lookForContiguousBlocks(cantBloques, fileData.InitialBlock, pathDialFS)
		if areFree {
			assignBlocks(bitmap, fileData.InitialBlock, cantBloques)
			showBitmap(bitmap)
			updateBitMap(bitmap, bitmapFilePath)
			updateMetaDataFile(pathDialFS, fileName, fileData.InitialBlock, length)
		} else {
			//log.Printf("No hay bloques contiguos disponibles")
			dataInBlock := getDataInBlockFile(blocksFilePath, fileData.InitialBlock, fileData.Size)
			//log.Printf("Data in block: %s", dataInBlock)
			truncateBitmap(bitmap, fileData.InitialBlock, bitmapFilePath, pathDialFS, fileName, fileData.Size, blocksFilePath)
			firstFreeBlock := firstBitFree(bitmap)
			assignBlocks(bitmap, firstFreeBlock, cantBloques)
			showBitmap(bitmap)
			updateBitMap(bitmap, bitmapFilePath)
			writeBlocksFile(blocksFilePath, firstFreeBlock, dataInBlock)
			updateMetaDataFile(pathDialFS, fileName, firstFreeBlock, length)
		}
	} else if length < fileData.Size {
		totalBlocks := (fileData.Size + config.TamanioBloqueDialFS - 1) / config.TamanioBloqueDialFS
		removeBlocks(bitmap, fileData.InitialBlock, totalBlocks, cantBloques)
		showBitmap(bitmap)
		updateBitMap(bitmap, bitmapFilePath)
		updateMetaDataFile(pathDialFS, fileName, fileData.InitialBlock, length)
	} else if length == fileData.Size {
		log.Printf("El tamaño a truncar es igual al tamaño actual del archivo")
	} else {
		log.Printf("Error al truncar el archivo, bloques disponibles %d", totalFreeBlocks)
	}
}

func assignBlocks(bitmap *Bitmap, initialBlock int, cantBloques int) {
	for i := initialBlock; i < initialBlock+cantBloques; i++ {
		bitmap.Set(i)
	}
}

func getTotalFreeBlocks(bitmap *Bitmap) int {
	totalFreeBlocks := 0
	for i := 0; i < config.CantidadBloquesDialFS; i++ {
		if !bitmap.Get(i) {
			totalFreeBlocks++
		}
	}
	return totalFreeBlocks
}

func removeBlocks(bitmap *Bitmap, initialBlock int, totalBlocks int, blocksToRemove int) {
	if blocksToRemove >= totalBlocks {
		// Remove all blocks if we're removing all or more than total
		for i := initialBlock; i < initialBlock+totalBlocks; i++ {
			bitmap.Remove(i)
		}
	} else {
		// Remove only the last 'blocksToRemove' blocks
		for i := initialBlock + totalBlocks - blocksToRemove; i < initialBlock+totalBlocks; i++ {
			bitmap.Remove(i)
		}
	}
}

func getDataInBlockFile(blocksFilePath string, initialBlock int, size int) string {
	// Abre el archivo bloques.dat para lectura
	file, err := os.OpenFile(blocksFilePath, os.O_RDWR, 0644)
	if err != nil {
		log.Fatalf("error al abrir el archivo bloques.dat: %v", err)
	}
	defer file.Close()

	var result strings.Builder

	var blocksToDelete int
	if size > config.TamanioBloqueDialFS {
		blocksToDelete = (size + config.TamanioBloqueDialFS - 1) / config.TamanioBloqueDialFS
	} else {
		blocksToDelete = 1
	}

	for i := initialBlock; i < initialBlock+blocksToDelete; i++ {
		posBloque := int64(i) * int64(config.TamanioBloqueDialFS)
		buffer := make([]byte, config.TamanioBloqueDialFS)

		// Lee el bloque en la posición calculada
		_, err := file.ReadAt(buffer, posBloque)
		if err != nil {
			if err == io.EOF {
				break // Fin del archivo
			}
			log.Fatalf("error al leer el bloque %d: %v", i, err)
		}

		// Añade el contenido del bloque al resultado
		result.Write(buffer)
	}

	return result.String()
}

func writeBlocksFile(blocksFilePath string, initialBlock int, dataInBlock string) {
	file, err := os.OpenFile(blocksFilePath, os.O_RDWR, 0644)
	if err != nil {
		log.Fatalf("error al abrir el archivo bloques.dat: %v", err)
	}
	defer file.Close()
	posicionInicialDeEscritura := (initialBlock * config.TamanioBloqueDialFS)
	//log.Printf("Posicion de escritura: %d", posicionInicialDeEscritura)

	// ME MUEVO A LA POSICION INICIAL DE ESCRITURA
	_, err = file.Seek(int64(posicionInicialDeEscritura), 0)
	if err != nil {
		log.Fatalf("Error al mover el cursor del archivo de bloques '%s': %v", blocksFilePath, err)
	}

	// ESCRIBIR EN EL ARCHIVO DE BLOQUES
	// Aquí deberías escribir los datos reales en lugar de "hola"
	// Por ejemplo, podrías usar los datos de 'adress' y 'length'
	dataToWrite := []byte(dataInBlock) // Reemplazar esto con los datos reales--->GLOBALmemoryContent
	_, err = file.Write(dataToWrite)
	if err != nil {
		log.Fatalf("Error al escribir en el archivo de bloques '%s': %v", blocksFilePath, err)
	}

}

func lookForContiguousBlocks(cantBloques int, initialBlock int, pathDialFS string) bool {
	// Abrir el archivo de bitmap para lectura
	bitmapFilePath := pathDialFS + "/bitmap.dat"

	// LEER EL CONTENIDO DEL ARCHIVO DE BITMAP
	bitmapBytes, err := os.ReadFile(bitmapFilePath)
	if err != nil {
		log.Fatalf("Error al leer el archivo de bitmap '%s': %v", bitmapFilePath, err)
	}

	// Crear un nuevo Bitmap y llenarlo con los datos leídos
	bitmap := NewBitmap()
	err = bitmap.FromBytes(bitmapBytes)
	if err != nil {
		log.Fatalf("Error al convertir bytes a bitmap: %v", err)
	}

	// Verificar si el rango está dentro de los límites del bitmap
	if initialBlock+cantBloques > config.CantidadBloquesDialFS {
		return false
	}

	// Verificar si todos los bloques en el rango están libres
	for i := initialBlock + 1; i < initialBlock+cantBloques; i++ {
		if bitmap.Get(i) {
			return false
		}
	}
	return true
}

func truncateBitmap(bitmap *Bitmap, initialBlock int, bitmapFilePath string, pathDialFS string, fileName string, fileSize int, blocksFilePath string) {
	//eliminamos los bloques que tiene asignado el archivo

	var blocksToDelete int
	if fileSize > config.TamanioBloqueDialFS {
		blocksToDelete = (fileSize + config.TamanioBloqueDialFS - 1) / config.TamanioBloqueDialFS
	} else {
		blocksToDelete = 1
	}
	removeBlocks(bitmap, initialBlock, blocksToDelete, blocksToDelete)
	deleteInBlockFile(pathDialFS, blocksToDelete, initialBlock)
	//log.Printf("Bloques eliminados de mi archivo: %d", blocksToDelete)
	//showBitmap(bitmap)
	deleteInMetaDataStructure(fileName)

	for _, fileContent := range metaDataStructure {
		//log.Printf("Archivo: %s ", fileContent.FileName)
		blocksPerFile := getBlocksFile(fileContent.FileName)
		//log.Printf("Bloques del archivo: %d", blocksPerFile)
		newInitialBlock := moveZeros(bitmap, fileContent.InitialBlock, blocksPerFile, bitmapFilePath, blocksFilePath)
		//log.Printf("Nuevo bloque inicial, de %s: %d \n", fileContent.FileName, newInitialBlock)
		updateMetaDataFile(pathDialFS, fileContent.FileName, newInitialBlock, fileContent.Size)

	}

}

func moveZeros(bitmap *Bitmap, initialBlock int, cantBloques int, bitmapFilePath string, blockFilePath string) int {
	snakeSize := 0
	newInitialBlock := initialBlock

	// Abre el archivo bloques.dat para lectura y escritura
	file, err := os.OpenFile(blockFilePath, os.O_RDWR, 0644)
	if err != nil {
		log.Fatalf("error al abrir el archivo bloques.dat: %v", err)
	}
	defer file.Close()

	for i := 0; i < initialBlock+cantBloques; i++ {
		if !bitmap.Get(i) {
			snakeSize++
		} else {
			if snakeSize > 0 {
				// Mover en el bitmap
				bitmap.Remove(i)
				bitmap.Set(i - snakeSize)

				// Mover en el archivo bloques.dat
				posBloque := int64(i) * int64(config.TamanioBloqueDialFS)
				newPos := int64(i-snakeSize) * int64(config.TamanioBloqueDialFS)

				buffer := make([]byte, config.TamanioBloqueDialFS)

				// Leer el bloque
				_, err := file.ReadAt(buffer, posBloque)
				if err != nil {
					log.Fatalf("error al leer el bloque %d: %v", i, err)
				}

				// Escribir el bloque en la nueva posición
				_, err = file.WriteAt(buffer, newPos)
				if err != nil {
					log.Fatalf("error al escribir el bloque %d en la posición %d: %v", i, i-snakeSize, err)
				}

				if i-snakeSize < newInitialBlock {
					newInitialBlock = i - snakeSize
				}
			}
		}
	}

	// Actualizar el bitmap
	updateBitMap(bitmap, bitmapFilePath)

	return newInitialBlock
}

/*
func removeBlocksFromFile(pathDialFS string, blocksToDelete int, initialBlock int) error {
	// Abre el archivo bloques.dat en modo lectura/escritura
	blocksFilePath := filepath.Join(pathDialFS, "bloques.dat")
	blocksFile, err := os.OpenFile(blocksFilePath, os.O_RDWR, 0644)
	if err != nil {
		log.Fatalf("Error al abrir el archivo de bloques '%s': %v", blocksFilePath, err)
	}
	defer blocksFile.Close()

	// Calcula la posición inicial para comenzar a eliminar
	startPosition := initialBlock * config.TamanioBloqueDialFS

	// Recorre la cantidad de bloques a eliminar
	for i := 0; i < blocksToDelete; i++ {
		// Mueve el cursor al bloque actual
		_, err := blocksFile.Seek(int64(startPosition+i*config.TamanioBloqueDialFS), 0)
		if err != nil {
			log.Fatalf("Error al mover el cursor del archivo de bloques '%s': %v", blocksFilePath, err)
		}

		// Escribe ceros en el bloque (o marca como libre)
		zeroBlock := make([]byte, config.TamanioBloqueDialFS)
		_, err = blocksFile.Write(zeroBlock)
		if err != nil {
			log.Fatalf("Error al escribir en el archivo de bloques '%s': %v", blocksFilePath, err)
		}
	}

	// Mostrar el contenido completo del archivo
	err = displayFileContent(blocksFilePath)
	if err != nil {
		log.Fatalf("Error al mostrar el contenido del archivo '%s': %v", blocksFilePath, err)
	}

	return nil
}

func displayFileContent(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Error al abrir el archivo '%s': %v", filePath, err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Error al leer el contenido del archivo '%s': %v", filePath, err)
	}

	fmt.Println("Contenido del archivo bloques.dat:")
	fmt.Println(string(content))

	return nil
}*/

func getBlocksFile(fileName string) int {
	var blocksToDelete int
	for _, fileContent := range metaDataStructure {
		if fileContent.FileName == fileName {
			if fileContent.Size > config.TamanioBloqueDialFS {
				blocksToDelete = (fileContent.Size + config.TamanioBloqueDialFS - 1) / config.TamanioBloqueDialFS
				return blocksToDelete
			} else {
				blocksToDelete = 1
				return blocksToDelete
			}
		}
	}
	return -1
}

/* ----------------------------------------- FUNCIONES DE FS_WRITE ------------------------------------------------------ */

func IO_FS_WRITE(pathDialFS string, fileName string, adress []int, length int, regPuntero int, pid int) {
	// VERIFICO EXISTENCIA DE ARCHIVO
	verificarExistenciaDeArchivo(pathDialFS, fileName)

	req := MemoryRequest{
		PID:     pid,
		Address: adress,
		Size:    length,
		Type:    "IO",
		Port:    config.Puerto,
	}

	err := SendAdressToMemory(req)
	if err != nil {
		log.Fatalf("Error al leer desde la memoria: %v", err)
	}

	// VERIFICO EXISTENCIA DE ARCHIVO
	verificarExistenciaDeArchivo(pathDialFS, fileName)

	// TENGO QUE ABRIR EL ARCHIVO DE BLOQUES.DAT
	blocksFilePath := filepath.Join(pathDialFS, "bloques.dat")
	blocksFile, err := os.OpenFile(blocksFilePath, os.O_RDWR, 0644)
	if err != nil {
		log.Fatalf("Error al abrir el archivo de bloques '%s': %v", blocksFilePath, err)
	}
	defer blocksFile.Close()

	// BLOQUE A ESCRIBIR
	bloqueInicialDelArchivo := firstBlockOfFileInMetadata(fileName)
	//fileData := dataFileInMetaDataStructure(fileName)

	posicionInicialDeEscritura := (bloqueInicialDelArchivo * config.TamanioBloqueDialFS) + regPuntero
	//log.Printf("Posicion de escritura: %d", posicionInicialDeEscritura)

	// ME MUEVO A LA POSICION INICIAL DE ESCRITURA
	_, err = blocksFile.Seek(int64(posicionInicialDeEscritura), 0)
	if err != nil {
		log.Fatalf("Error al mover el cursor del archivo de bloques '%s': %v", blocksFilePath, err)
	}

	// ESCRIBIR EN EL ARCHIVO DE BLOQUES
	// Aquí deberías escribir los datos reales en lugar de "hola"
	// Por ejemplo, podrías usar los datos de 'adress' y 'length'
	dataToWrite := []byte(GLOBALmemoryContent) // Reemplazar esto con los datos reales--->GLOBALmemoryContent
	_, err = blocksFile.Write(dataToWrite)
	if err != nil {
		log.Fatalf("Error al escribir en el archivo de bloques '%s': %v", blocksFilePath, err)
	}

}

/* ------------------------------------------ FUNCIONES DE FS_READ ------------------------------------------------------ */

func IO_FS_READ(pathDialFS string, fileName string, address []int, length int, regPuntero int, pid int) {
	// VERIFICO EXISTENCIA DE ARCHIVO
	verificarExistenciaDeArchivo(pathDialFS, fileName)

	// TENGO QUE ABRIR EL ARCHIVO DE BLOQUES.DAT
	blocksFilePath := pathDialFS + "/bloques.dat"
	blocksFile, err := os.Open(blocksFilePath)
	if err != nil {
		log.Fatalf("Error al abrir el archivo de bloques '%s': %v", blocksFilePath, err)
	}
	defer blocksFile.Close()

	// BLOQUE A LEER
	bloqueInicialDelArchivo := firstBlockOfFileInMetadata(fileName)
	posicionInicialDeLectura := (bloqueInicialDelArchivo * config.TamanioBloqueDialFS) + regPuntero

	// ME MUEVO A LA POSICION INICIAL DE LECTURA
	_, err = blocksFile.Seek(int64(posicionInicialDeLectura), 0)
	if err != nil {
		log.Fatalf("Error al mover el cursor del archivo de bloques '%s': %v", blocksFilePath, err)
	}

	// LEO LA CANTIDAD DE BYTES INDICADA POR LENGTH Y CREO UN SLICE CON UN TAMAÑO DEFINIDO PARA ALMACERNARLO
	contenidoLeidoDeArchivo := make([]byte, length)
	_, err = blocksFile.Read(contenidoLeidoDeArchivo)
	if err != nil {
		log.Fatalf("Error al leer el archivo de bloques '%s': %v", blocksFilePath, err)
	}

	// TENGO QUE ESCRIBIR EL CONTENIDO LEIDO EN MEMORIA A PARTIR DE LA DIRECCION FISICA INDICADA EN ADDRESS
	// Llamo a endpoint para escribir el contenido en memoria

	err1 := SendInputToMemory(pid, string(contenidoLeidoDeArchivo), address)
	if err1 != nil {
		log.Fatalf("Error al leer desde la memoria: %v", err)
	}
}

/* ------------------------------------- CREAR ARCHIVOS DE BLOQUES Y BITMAP ------------------------------------------------------ */

func CreateBlockFile(path string, blocksSize int, blocksCount int, sizeFile int) (*BlockFile, error) {

	filePath := path + "/bloques.dat"

	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Error al crear el archivo '%s': %v", path, err)
	}
	defer file.Close()

	// ASIGNO EL TAMAÑO DEL ARCHIVO AL QUE DICE EL CONFIG
	err = file.Truncate(int64(sizeFile))
	if err != nil {
		log.Fatalf("Error al truncar el archivo '%s': %v", path, err)
	}

	return &BlockFile{
		FilePath:    filePath,
		BlocksSize:  blocksSize,
		BlocksCount: blocksCount,
		FreeBlocks:  make([]bool, blocksCount),
	}, nil
}

func CreateBitmapFile(path string, blocksCount int, bitmapSize int) {
	filePath := path + "/bitmap.dat"

	bitmapFile, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Error al crear el archivo de bitmap '%s': %v", filePath, err)
	}
	defer bitmapFile.Close()

	bitmap := NewBitmap()

	bitmapBytes := bitmap.ToBytes()
	_, err = bitmapFile.Write(bitmapBytes)
	if err != nil {
		log.Fatalf("Error al inicializar el archivo de bitmap '%s': %v", filePath, err)
	}

	// flushear si hubo error
	if err := bitmapFile.Sync(); err != nil {
		log.Fatalf("Error al forzar la escritura del archivo de bitmap '%s': %v", filePath, err)
	}
}

func EnsureIfFileExists(pathDialFS string, blocksSize int, blocksCount int, sizeFile int, bitmapSize int) {

	// pathDialFS completa para bloques.dat
	blockFilePath := pathDialFS + "/bloques.dat"
	if _, err := os.Stat(blockFilePath); os.IsNotExist(err) {
		CreateBlockFile(pathDialFS, blocksSize, blocksCount, sizeFile)
	}

	// pathDialFS completa para bitmap.dat
	bitmapFilePath := pathDialFS + "/bitmap.dat"
	if _, err := os.Stat(bitmapFilePath); os.IsNotExist(err) {
		CreateBitmapFile(pathDialFS, blocksCount, bitmapSize)
	}
}

func verificarExistenciaDeArchivo(path string, fileName string) {
	filePath := path + "/" + fileName
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Fatalf("El archivo '%s' no existe", fileName)
	}
}

/* --------------------------------------------- METODOS DEL BITMAP ------------------------------------------------------ */

func NewBitmap() *Bitmap {
	return &Bitmap{
		bits:       make([]int, config.CantidadBloquesDialFS),
		blockCount: config.CantidadBloquesDialFS,
		blockSize:  config.TamanioBloqueDialFS,
	}
}

func (b *Bitmap) FromBytes(bytes []byte) error {
	expectedBytes := (b.blockCount * b.blockSize) / 8
	if len(bytes) != expectedBytes {
		return fmt.Errorf("invalid byte slice length: expected %d bytes, got %d", expectedBytes, len(bytes))
	}

	b.bits = make([]int, b.blockCount)
	for i := 0; i < b.blockCount; i++ {
		byteIndex := (i * b.blockSize) / 8
		bitOffset := (i * b.blockSize) % 8
		if bytes[byteIndex]&(1<<bitOffset) != 0 {
			b.bits[i] = 1
		}
	}
	return nil
}

func (b *Bitmap) ToBytes() []byte {
	bytes := make([]byte, (b.blockCount*b.blockSize)/8)
	for i := 0; i < b.blockCount; i++ {
		if b.bits[i] == 1 {
			byteIndex := (i * b.blockSize) / 8
			bitOffset := (i * b.blockSize) % 8
			bytes[byteIndex] |= 1 << bitOffset
		}
	}
	return bytes
}

func (b *Bitmap) Get(pos int) bool {
	if pos < 0 || pos >= b.blockCount {
		return false
	}
	return b.bits[pos] == 1
}

func (b *Bitmap) Set(pos int) {
	if pos < 0 || pos >= b.blockCount {
		return
	}
	b.bits[pos] = 1
}

func (b *Bitmap) Remove(pos int) {
	if pos < 0 || pos >= b.blockCount {
		return
	}
	b.bits[pos] = 0
}

/* ------------------------------------- FUNCIONES PARA AGILIZAR BITMAP ------------------------------------------------------ */

func showBitmap(bitmap *Bitmap) {
	fmt.Println("Bitmap:")
	for i := 0; i < config.CantidadBloquesDialFS; i++ {
		if bitmap.Get(i) {
			fmt.Print("1")
		} else {
			fmt.Print("0")
		}
		if (i+1)%64 == 0 {
			fmt.Println() // New line every 64 bits for readability
		}
	}
}

func readAndCopyBitMap(bitmapFilePath string) *Bitmap {
	bitmapBytes, err := os.ReadFile(bitmapFilePath)
	if err != nil {
		log.Fatalf("Error al leer el archivo de bitmap '%s': %v", bitmapFilePath, err)
	}

	bitmap := NewBitmap()
	err = bitmap.FromBytes(bitmapBytes)
	if err != nil {
		log.Fatalf("Error al convertir bytes a bitmap: %v", err)
	}

	return bitmap
}

func updateBitMap(bitmap *Bitmap, bitmapFilePath string) {
	modifiedBitmapBytes := bitmap.ToBytes()

	err := os.WriteFile(bitmapFilePath, modifiedBitmapBytes, 0644)
	if err != nil {
		log.Fatalf("Error al escribir el archivo de bitmap modificado '%s': %v", bitmapFilePath, err)
	}

}

/* --------------------------------------------- FUNCIONES DE METADATA ------------------------------------------------------ */

func updateMetaDataFile(pathDialFS string, fileName string, initialBlock int, fileSize int) {
	filePath := pathDialFS + "/" + fileName
	fileContent := FileContent{
		InitialBlock: initialBlock,
		Size:         fileSize,
	}
	contentBytes, err := json.Marshal(fileContent)
	if err != nil {
		log.Fatalf("Error al convertir FileContent a bytes: %v", err)
	}

	// Write FileContent to the file
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		log.Fatalf("Error al crear el archivo '%s': %v", pathDialFS, err)
	}
	_, err = file.Write(contentBytes)
	if err != nil {
		log.Fatalf("Error al escribir el contenido en el archivo '%s': %v", filePath, err)
	}

	file.Close()
}

func deleteInMetaDataStructure(fileName string) {
	for i, fileContent := range metaDataStructure {
		if fileContent.FileName == fileName {
			metaDataStructure = append(metaDataStructure[:i], metaDataStructure[i+1:]...)
			break
		}
	}
}

func ensureExistingMetaDataFiles(pathDialFS string) {
	if checkFilesInDirectoryThatWereInDirectory(pathDialFS) {
		// Example usage of readFilesThatWereInDirectory
		metaDataStructure = readFilesThatWereInDirectory(pathDialFS)

		// Display filesContent
		/*for i, fileContent := range metaDataStructure {
			fmt.Printf("MetaStructure %d: FileName %s InitialBlock: %d, Size: %d\n", i, fileContent.FileName, fileContent.InitialBlock, fileContent.Size)
		}*/
	}
}

func readFilesThatWereInDirectory(directoryPath string) []FileContent {
	var filesContent []FileContent

	files, err := os.ReadDir(directoryPath)
	if err != nil {
		log.Fatalf("Error al leer el directorio '%s': %v", directoryPath, err)
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".txt" {
			filePath := filepath.Join(directoryPath, file.Name())
			fileContent := readFileOfMetaData(filePath)
			filesContent = append(filesContent, fileContent)
		}
	}
	sort.Slice(filesContent, func(i, j int) bool {
		return filesContent[i].InitialBlock < filesContent[j].InitialBlock
	})

	return filesContent
}

func checkFilesInDirectoryThatWereInDirectory(pathDialFS string) bool {
	files, err := os.ReadDir(pathDialFS)
	if err != nil {
		log.Printf("Error reading directory: %v", err)
		return false
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".txt") {
			return true
		}
	}
	return false
}

func readFileOfMetaData(pathFile string) FileContent {
	readContent, err := os.ReadFile(pathFile)
	if err != nil {
		log.Fatalf("Error al leer el archivo '%s': %v", pathFile, err)
	}

	var fileContent FileContent
	err = json.Unmarshal(readContent, &fileContent)
	if err != nil {
		log.Fatalf("Error al deserializar el contenido del archivo '%s': %v", pathFile, err)
	}

	fileContent.FileName = filepath.Base(pathFile)

	return fileContent
}

func firstBlockOfFileInMetadata(fileName string) int {
	for _, fileContent := range metaDataStructure {
		if fileContent.FileName == fileName {
			return fileContent.InitialBlock
		}
	}
	return -1
}

func dataFileInMetaDataStructure(fileName string) (FileContent, error) {
	for _, fileContent := range metaDataStructure {
		if fileContent.FileName == fileName {
			return fileContent, nil
		}
	}
	return FileContent{}, fmt.Errorf("file '%s' not found in metadata structure", fileName)
}
