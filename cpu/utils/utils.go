package utils

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"
)

/*---------------------------------------------- STRUCTS --------------------------------------------------------*/
type KernelRequest struct {
	PcbUpdated     PCB    `json:"pcbUpdated"`
	MotivoDesalojo string `json:"motivoDesalojo"`
	TimeIO         int    `json:"timeIO"`
	Interface      string `json:"interface"`
	IoType         string `json:"ioType"`
	Recurso        string `json:"recurso"`
}

type PCB struct { //ESTO NO VA ACA
	Pid, Quantum int
	State        string
	CpuReg       RegisterCPU
}

type ExecutionContext struct {
	Pid    int
	State  string
	CpuReg RegisterCPU
}

type RegisterCPU struct {
	PC, EAX, EBX, ECX, EDX, SI, DI uint32
	AX, BX, CX, DX                 uint8
}

type BodyResponseInstruction struct {
	Instruction string `json:"instruction"`
}

type ResponseInterrupt struct {
	Interrupt bool   `json:"interrupt"`
	Pid       int    `json:"pid"`
	Motivo    string `json:"motivo"`
}

type ResponseWait struct {
	Recurso string `json:"recurso"`
	Pid     int    `json:"pid"`
}

type TranslationRequest struct {
	DireccionLogica int `json:"logical_address"`
	TamPag          int `json:"page_size"`
	TamData         int `json:"data_size"`
	PID             int `json:"pid"`
}

type TranslationResponse struct {
	DireccionesFisicas []int `json:"physical_addresses"`
}

type TLBEntry struct {
	PID                int
	Pagina             int
	Frame              int
	UltimoAcceso       time.Time // Para LRU
	globalPosicionFila int       // Para FIFO
}

type bodyProcess struct {
	Pid   int `json:"pid"`
	Pages int `json:"pages,omitempty"`
}

type bodyPageTable struct {
	Pid  int `json:"pid"`
	Page int `json:"page"`
}

type BodyFrame struct {
	Frame int `json:"frame"`
}
type bodyRegisters struct {
	Pid       int   `json:"iopid"`
	DirFisica []int `json:"dirFisica"`
	LengthREG int   `json:"lengthREG"`
}

type BodyPageTam struct {
	PageTam int `json:"pageTam"`
}

type BodyContent struct {
	Content string `json:"content"`
}

type MemoryReadRequest struct {
	PID     int    `json:"pid"`
	Address []int  `json:"address"`
	Size    int    `json:"size,omitempty"` //Si es 0, se omite (Util para creacion y terminacion de procesos)
	Data    []byte `json:"data,omitempty"` //Si es 0, se omite Util para creacion y terminacion de procesos)
	Type    string `json:"type"`
}

type FSstructure struct {
	FileName      string `json:"filename"`
	FSInstruction string `json:"fsinstruction"`
	FSRegTam      int    `json:"fsregtam"`
	FSRegDirec    []int  `json:"fsregdirec"`
	FSRegPuntero  int    `json:"fsregpuntero"`
}

/*------------------------------------------------- VAR GLOBALES --------------------------------------------------------*/

var globalTLB []TLBEntry
var globalTLBsize int
var replacementAlgorithm string
var globalPosicionFila int
var interrupt bool = false
var GLOBALrequestCPU KernelRequest
var GLOBALcontextoDeEjecucion PCB //PCB recibido desde kernel
var MemoryFrame int
var GLOBALpageTam int
var GLOBALdataMOV_IN []byte

// var requestCPU KernelRequest
var responseInterruptGlobal ResponseInterrupt

/*func init() {
	globals.ClientConfig = IniciarConfiguracion(os.Args[1]) // tiene que prender la confi cuando arranca
}*/

func ConfigurarLogger() {

	logFile, err := os.OpenFile("cpu.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
}

func init() {
	globals.ClientConfig = IniciarConfiguracion(os.Args[1]) // tiene que prender la confi cuando arranca

	if globals.ClientConfig != nil {
		globalTLBsize = globals.ClientConfig.NumberFellingTLB
		replacementAlgorithm = globals.ClientConfig.AlgorithmTLB
	} else {
		log.Fatal("ClientConfig is not initialized")
	}
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

func ReceivePCB(w http.ResponseWriter, r *http.Request) {
	// HAGO UN LOG PARA CHEQUEAR RECEPCION

	err := json.NewDecoder(r.Body).Decode(&GLOBALcontextoDeEjecucion)
	if err != nil {
		http.Error(w, "Error al decodificar los datos JSON", http.StatusInternalServerError)
		return
	}

	InstructionCycle(GLOBALcontextoDeEjecucion)
	w.WriteHeader(http.StatusOK)
}

func InstructionCycle(contextoDeEjecucion PCB) {
	GLOBALrequestCPU = KernelRequest{}

	for {
		log.Printf("PID: %d - FETCH - Program Counter: %d\n", contextoDeEjecucion.Pid, contextoDeEjecucion.CpuReg.PC)
		line, _ := Fetch(int(contextoDeEjecucion.CpuReg.PC), contextoDeEjecucion.Pid)

		contextoDeEjecucion.CpuReg.PC++
		GLOBALrequestCPU.PcbUpdated = contextoDeEjecucion
		instruction, _ := Decode(line)

		log.Printf("PID: %d - Ejecutando: %s - %s.", contextoDeEjecucion.Pid, instruction, line)
		Execute(instruction, line, &contextoDeEjecucion)

		if (responseInterruptGlobal.Interrupt && responseInterruptGlobal.Pid == contextoDeEjecucion.Pid) || interrupt {
			responseInterruptGlobal.Interrupt = false
			interrupt = false
			break
		}

	}
	if GLOBALrequestCPU.MotivoDesalojo == "" {
		GLOBALrequestCPU.MotivoDesalojo = responseInterruptGlobal.Motivo
	}
	GLOBALrequestCPU.PcbUpdated = contextoDeEjecucion
	responsePCBtoKernel(GLOBALrequestCPU)

}

func responsePCBtoKernel(requestCPU KernelRequest) {
	kernelURL := fmt.Sprintf("http://%s:%d/syscall", globals.ClientConfig.IpKernel, globals.ClientConfig.PortKernel)

	requestJSON, err := json.Marshal(requestCPU)
	if err != nil {
		return
	}
	resp, err := http.Post(kernelURL, "application/json", bytes.NewBuffer(requestJSON))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}
}

func Fetch(pc int, pid int) ([]string, error) {
	memoriaURL := fmt.Sprintf("http://%s:%d/getInstructionFromPid?pid=%d&programCounter=%d", globals.ClientConfig.IPMemory, globals.ClientConfig.PortMemory, pid, pc)
	resp, err := http.Get(memoriaURL)
	if err != nil {
		log.Fatalf("error al enviar la solicitud al módulo de memoria: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("error en la respuesta del módulo de memoria: %v", resp.StatusCode)
		log.Println(err)
		return nil, err
	}

	var response BodyResponseInstruction
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Println("error al decodificar la respuesta del módulo de memoria:", err)
		return nil, err
	}

	instructions := strings.Split(response.Instruction, ",") // split the string into a slice
	return instructions, nil
}

func Decode(instruction []string) (string, error) {
	// Esta función se va a complejizar con la traducción de las direcciones fisicas y logicas
	words := strings.Fields(instruction[0])
	if len(instruction) == 0 {
		return "nil", fmt.Errorf("instrucción vacía")
	}
	return words[0], nil
}

func Execute(instruction string, line []string, contextoDeEjecucion *PCB) error {

	words := strings.Fields(line[0])

	switch instruction {
	case "SET": // Change the type of the switch case expression from byte to string
		err := SetCampo(&contextoDeEjecucion.CpuReg, words[1], words[2])
		if err != nil {
			return fmt.Errorf("error en execute: %s", err)
		}
	case "SUM":
		err := Suma(&contextoDeEjecucion.CpuReg, words[1], words[2])
		if err != nil {
			return fmt.Errorf("error en execute: %s", err)
		}
	case "SUB":
		err := Resta(&contextoDeEjecucion.CpuReg, words[1], words[2])
		if err != nil {
			return fmt.Errorf("error en execute: %s", err)
		}
	case "JNZ":
		err := JNZ(&contextoDeEjecucion.CpuReg, words[1], words[2])
		if err != nil {
			return fmt.Errorf("error en execute: %s", err)
		}
	case "IO_GEN_SLEEP":
		err := IO(instruction, words, contextoDeEjecucion)
		if err != nil {
			return fmt.Errorf("error en execute: %s", err)
		}
	case "IO_STDIN_READ":
		err := IO(instruction, words, contextoDeEjecucion)
		if err != nil {
			return fmt.Errorf("error en execute: %s", err)
		}
	case "IO_STDOUT_WRITE":
		err := IO(instruction, words, contextoDeEjecucion)
		if err != nil {
			return fmt.Errorf("error en execute: %s", err)
		}
	case "RESIZE":
		tam, err := strconv.Atoi(words[1])
		if err != nil {
			return fmt.Errorf("error en execute: %s", err)
		}
		sendResizeMemory(tam)

	case "MOV_IN":
		err := MOV_IN(words, contextoDeEjecucion)
		if err != nil {
			return fmt.Errorf("error en execute: %s", err)
		}
	case "MOV_OUT":
		err := MOV_OUT(words, contextoDeEjecucion)
		if err != nil {
			return fmt.Errorf("error en execute: %s", err)
		}
	case "COPY_STRING":
		err := COPY_STRING(words, contextoDeEjecucion)
		if err != nil {
			return fmt.Errorf("error en execute: %s", err)
		}
	case "WAIT":
		err := CheckWait(nil, nil, contextoDeEjecucion, words[1])
		if err != nil {
			return fmt.Errorf("error en execute: %s", err)
		}
	case "SIGNAL":
		err := CheckSignal(nil, nil, contextoDeEjecucion.Pid, instruction, words[1])
		if err != nil {
			return fmt.Errorf("error en execute: %s", err)

		}
	case "IO_FS_CREATE":
		err := IO(instruction, words, contextoDeEjecucion)
		if err != nil {
			return fmt.Errorf("error en execute: %s", err)
		}
	case "IO_FS_DELETE":
		err := IO(instruction, words, contextoDeEjecucion)
		if err != nil {
			return fmt.Errorf("error en execute: %s", err)
		}
	case "IO_FS_TRUNCATE":
		err := IO(instruction, words, contextoDeEjecucion)
		if err != nil {
			return fmt.Errorf("error en execute: %s", err)
		}
	case "IO_FS_WRITE":
		err := IO(instruction, words, contextoDeEjecucion)
		if err != nil {
			return fmt.Errorf("error en execute: %s", err)
		}
	case "IO_FS_READ":
		err := IO(instruction, words, contextoDeEjecucion)
		if err != nil {
			return fmt.Errorf("error en execute: %s", err)
		}
	case "EXIT":
		err := TerminarProceso(&contextoDeEjecucion.CpuReg, "FINALIZADO")
		if err != nil {
			return fmt.Errorf("error en execute: %s", err)
		}
	default:
		return nil
	}
	return nil
}

func TerminarProceso(registerCPU *RegisterCPU, motivo string) error {
	GLOBALrequestCPU = KernelRequest{
		MotivoDesalojo: motivo,
	}

	interrupt = true // Aquí va el valor booleano que quieres enviar al kernel
	registerCPU.PC--
	return nil
}

func SetCampo(r *RegisterCPU, campo string, valor interface{}) error {
	// Obtener el valor reflect.Value de la estructura Persona
	valorRef := reflect.ValueOf(r)

	// Obtener el campo especificado por el parámetro campo
	campoRef := valorRef.Elem().FieldByName(campo)

	// Verificar si el campo existe
	if !campoRef.IsValid() {
		return fmt.Errorf("campo '%s' no encontrado en la estructura", campo)
	}

	// Obtener el tipo de dato del campo
	tipoCampo := campoRef.Type()

	// Convertir el valor proporcionado al tipo de dato del campo
	switch tipoCampo.Kind() {
	case reflect.String:
		campoRef.SetString(fmt.Sprintf("%v", valor))
	case reflect.Uint8:
		valorUint, err := strconv.ParseUint(fmt.Sprintf("%v", valor), 10, 8)
		if err != nil {
			return err
		}
		campoRef.SetUint(valorUint)
	case reflect.Uint32:
		valorUint, err := strconv.ParseUint(fmt.Sprintf("%v", valor), 10, 32)
		if err != nil {
			return err
		}
		campoRef.SetUint(valorUint)
	case reflect.Int:
		valorInt, err := strconv.ParseInt(fmt.Sprintf("%v", valor), 10, 64)
		if err != nil {
			return err
		}
		campoRef.SetInt(valorInt)
	// Agrega más casos según sea necesario para otros tipos de datos
	default:
		return fmt.Errorf("tipo de dato del campo '%s' no soportado", tipoCampo)
	}

	return nil
}

func Suma(registerCPU *RegisterCPU, s1, s2 string) error {
	valorRef := reflect.ValueOf(registerCPU)

	campoDestinoRef := valorRef.Elem().FieldByName(s1)

	if !campoDestinoRef.IsValid() {
		return fmt.Errorf("campo destino '%s' no encontrado en la estructura", s1)
	}

	tipoCampoDestino := campoDestinoRef.Type()

	campoOrigenRef := valorRef.Elem().FieldByName(s2)

	if !campoOrigenRef.IsValid() {
		return fmt.Errorf("campo origen '%s' no encontrado en la estructura", s2)
	}

	tipoCampoOrigen := campoOrigenRef.Type()
	if tipoCampoDestino != tipoCampoOrigen {
		return fmt.Errorf("los campos '%s' y '%s' no son del mismo tipo", s1, s2)
	}

	switch tipoCampoDestino.Kind() {
	case reflect.Uint8:
		valorDestino := campoDestinoRef.Uint()
		valorOrigen := campoOrigenRef.Uint()
		suma := valorDestino + valorOrigen
		if suma > 255 {
			suma = 255
		}
		campoDestinoRef.SetUint(suma)
	case reflect.Uint32:
		valorDestino := campoDestinoRef.Uint()
		valorOrigen := campoOrigenRef.Uint()
		suma := valorDestino + valorOrigen
		if suma > 4294967295 { // Máximo valor para uint32
			suma = 4294967295
		}
		campoDestinoRef.SetUint(suma)
	}
	return nil
}

func Resta(registerCPU *RegisterCPU, s1, s2 string) error {
	valorRef := reflect.ValueOf(registerCPU)

	campoDestinoRef := valorRef.Elem().FieldByName(s1)
	if !campoDestinoRef.IsValid() {
		return fmt.Errorf("campo destino '%s' no encontrado en la estructura", s1)
	}

	tipoCampoDestino := campoDestinoRef.Type()

	campoOrigenRef := valorRef.Elem().FieldByName(s2)
	if !campoOrigenRef.IsValid() {
		return fmt.Errorf("campo origen '%s' no encontrado en la estructura", s2)
	}

	tipoCampoOrigen := campoOrigenRef.Type()
	if tipoCampoDestino != tipoCampoOrigen {
		return fmt.Errorf("los campos '%s' y '%s' no son del mismo tipo", s1, s2)
	}

	switch tipoCampoDestino.Kind() {
	case reflect.Uint8:
		valorDestino := campoDestinoRef.Uint()
		valorOrigen := campoOrigenRef.Uint()
		resta := valorDestino - valorOrigen
		if resta <= 0 {
			resta = 0
		}
		campoDestinoRef.SetUint(resta)
	case reflect.Uint32:
		valorDestino := campoDestinoRef.Uint()
		valorOrigen := campoOrigenRef.Uint()
		if valorDestino < valorOrigen {
			campoDestinoRef.SetUint(0)
		} else {
			resta := valorDestino - valorOrigen
			campoDestinoRef.SetUint(resta)
		}
	}
	return nil
}

func JNZ(registerCPU *RegisterCPU, reg, valor string) error {
	// Obtener el valor reflect.Value de la estructura RegisterCPU
	valorRef := reflect.ValueOf(registerCPU)

	// Obtener el valor reflect.Value del campo destino
	campoDestinoRef := valorRef.Elem().FieldByName(reg)

	if !campoDestinoRef.IsValid() {
		return fmt.Errorf("campo destino '%s' no encontrado en la estructura", reg)
	}

	// Obtener el valor del campo destino
	campoDestinoValor := campoDestinoRef.Uint()

	if campoDestinoValor != 0 {
		valorUint32, err := strconv.ParseUint(valor, 10, 32)
		if err != nil {
			return err
		}
		registerCPU.PC = uint32(valorUint32)
	}

	return nil
}

func MOV_IN(words []string, contextoEjecucion *PCB) error {
	REGdireccion := words[2]
	valueDireccion := verificarRegistro(REGdireccion, contextoEjecucion)

	REGdatos := words[1]

	// Verificar el tipo de dato del registro en RegisterCPU
	var tamREGdatos int
	switch REGdatos {
	case "PC", "EAX", "EBX", "ECX", "EDX", "SI", "DI":
		tamREGdatos = 4 // uint32
	case "AX", "BX", "CX", "DX":
		tamREGdatos = 1 // uint8
	default:
		return fmt.Errorf("registro no soportado: %s", REGdatos)
	}

	direcciones := TranslateAddress(contextoEjecucion.Pid, valueDireccion, GLOBALpageTam, tamREGdatos)

	err1 := LeerMemoria(contextoEjecucion.Pid, direcciones, tamREGdatos)
	if err1 != nil {
		return fmt.Errorf("error leyendo memoria: %s", err1)
	}

	//buf := bytes.NewReader(GLOBALdataMOV_IN)
	if tamREGdatos == 1 {
		result := stringToUint8(string(GLOBALdataMOV_IN))
		log.Printf("PID: %d - Acción: LEER - Dirección Física: %d - Valor: %d", contextoEjecucion.Pid, direcciones[0], result)
		err3 := SetCampo(&contextoEjecucion.CpuReg, REGdatos, result)
		if err3 != nil {
			return fmt.Errorf("error en execute: %s", err3)
		}
	} else {
		result := stringToInteger(string(GLOBALdataMOV_IN))
		log.Printf("PID: %d - Acción: LEER - Dirección Física: %d - Valor: %d", contextoEjecucion.Pid, direcciones[0], result)
		err3 := SetCampo(&contextoEjecucion.CpuReg, REGdatos, result)
		if err3 != nil {
			return fmt.Errorf("error en execute: %s", err3)
		}
	}
	return nil
}

func stringToInteger(s string) uint32 {
	var result uint32 = 0
	for i, char := range s {
		// Convertimos cada carácter a su valor ASCII y lo desplazamos
		result |= uint32(char) << (8 * (3 - i))
		if i == 3 {
			break // Solo usamos los primeros 4 caracteres
		}
	}
	return result
}

func stringToUint8(s string) uint8 {
	if len(s) == 0 {
		return 0
	}
	// Tomamos solo el primer carácter del string
	return uint8(s[0])
}

func MOV_OUT(words []string, contextoEjecucion *PCB) error {
	REGdireccion := words[1]
	valueDireccion := verificarRegistro(REGdireccion, contextoEjecucion)

	REGdatos := words[2]
	valueDatos := verificarRegistro(REGdatos, contextoEjecucion)

	var valueDatosBytes []byte
	switch REGdatos {
	case "PC", "EAX", "EBX", "ECX", "EDX", "SI", "DI":
		valueDatosBytes = make([]byte, 4)
		binary.BigEndian.PutUint32(valueDatosBytes, uint32(valueDatos))
	case "AX", "BX", "CX", "DX":
		valueDatosBytes = []byte{uint8(valueDatos)}
	default:
		return fmt.Errorf("registro no soportado: %s", REGdatos)
	}
	direcciones := TranslateAddress(contextoEjecucion.Pid, valueDireccion, GLOBALpageTam, len(valueDatosBytes))
	err := EscribirMemoria(contextoEjecucion.Pid, direcciones, valueDatosBytes)
	if err != nil {
		return err
	}
	log.Printf("PID: %d - Acción: ESCRIBIR - Dirección Física: %d - Valor: %d", contextoEjecucion.Pid, direcciones[0], valueDatos)
	return nil
}

func COPY_STRING(words []string, contextoEjecucion *PCB) error {
	tamString := words[1]
	tam, err := strconv.Atoi(tamString)
	if err != nil {
		return err
	}
	valorSI := verificarRegistro("SI", contextoEjecucion)
	direccionesSI := TranslateAddress(contextoEjecucion.Pid, valorSI, GLOBALpageTam, tam)

	err1 := LeerMemoria(contextoEjecucion.Pid, direccionesSI, tam)
	if err1 != nil {
		return err1
	}
	log.Printf("PID: %d - Acción: LEER - Dirección Física: %v - Valor: %s", contextoEjecucion.Pid, direccionesSI[0], GLOBALdataMOV_IN)
	valorDI := verificarRegistro("DI", contextoEjecucion)
	direccionesDI := TranslateAddress(contextoEjecucion.Pid, valorDI, GLOBALpageTam, tam)
	err2 := EscribirMemoria(contextoEjecucion.Pid, direccionesDI, GLOBALdataMOV_IN)
	if err2 != nil {
		return err2
	}
	log.Printf("PID: %d - Acción: ESCRIBIR - Dirección Física: %d - Valor: %s", contextoEjecucion.Pid, direccionesDI[0], GLOBALdataMOV_IN)
	return nil
}

func LeerMemoria(pid int, direccion []int, size int) error {
	memoriaURL := fmt.Sprintf("http://%s:%d/readMemory", globals.ClientConfig.IPMemory, globals.ClientConfig.PortMemory)
	req := MemoryReadRequest{
		PID:     pid,
		Address: direccion,
		Size:    size,
		Type:    "CPU",
	}

	reqJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(memoriaURL, "application/json", bytes.NewBuffer(reqJSON))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error en la respuesta del módulo de memoria: %v", resp.StatusCode)
	}

	return nil
}

func RecieveMOV_IN(w http.ResponseWriter, r *http.Request) {
	var Content []byte
	err := json.NewDecoder(r.Body).Decode(&Content)
	if err != nil {
		http.Error(w, "Error al decodificar los datos JSON", http.StatusInternalServerError)
		return
	}

	GLOBALdataMOV_IN = Content
	w.WriteHeader(http.StatusOK)
}

func EscribirMemoria(pid int, direcciones []int, data []byte) error {
	memoriaURL := fmt.Sprintf("http://%s:%d/writeMemory", globals.ClientConfig.IPMemory, globals.ClientConfig.PortMemory)
	var req MemoryReadRequest
	req.PID = pid
	req.Address = direcciones
	req.Data = data

	reqJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(memoriaURL, "application/json", bytes.NewBuffer(reqJSON))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error en la respuesta del módulo de memoria: %v", resp.StatusCode)
	}

	return nil
}

func IO(kind string, words []string, contextoEjecucion *PCB) error {
	interrupt = true

	switch kind {
	case "IO_GEN_SLEEP":
		timeIO, err := strconv.Atoi(words[2])
		if err != nil {
			return err
		}
		GLOBALrequestCPU = KernelRequest{
			PcbUpdated:     *contextoEjecucion,
			MotivoDesalojo: "INTERRUPCION POR IO",
			IoType:         "GENERICA",
			Interface:      words[1],
			TimeIO:         timeIO,
		}
	case "IO_STDIN_READ":
		adressREG := words[2]
		valueAdress1 := verificarRegistro(adressREG, contextoEjecucion)

		lengthREG := words[3]
		valueLength1 := verificarRegistro(lengthREG, contextoEjecucion)

		direcciones := TranslateAddress(contextoEjecucion.Pid, valueAdress1, GLOBALpageTam, valueLength1)
		sendREGtoKernel(direcciones, valueLength1, contextoEjecucion.Pid)
		GLOBALrequestCPU = KernelRequest{
			PcbUpdated:     *contextoEjecucion,
			MotivoDesalojo: "INTERRUPCION POR IO",
			IoType:         "STDIN",
			Interface:      words[1],
			TimeIO:         0,
		}
	case "IO_STDOUT_WRITE":
		adressREG := words[2]
		valueAdress := verificarRegistro(adressREG, contextoEjecucion)

		lengthREG := words[3]
		valueLength := verificarRegistro(lengthREG, contextoEjecucion)

		direcciones := TranslateAddress(contextoEjecucion.Pid, valueAdress, GLOBALpageTam, valueLength)
		sendREGtoKernel(direcciones, valueLength, contextoEjecucion.Pid)
		GLOBALrequestCPU = KernelRequest{
			PcbUpdated:     *contextoEjecucion,
			MotivoDesalojo: "INTERRUPCION POR IO",
			IoType:         "STDOUT",
			Interface:      words[1],
			TimeIO:         0,
		}
	case "IO_FS_CREATE":
		fileName := words[2]
		GLOBALrequestCPU = KernelRequest{
			PcbUpdated:     *contextoEjecucion,
			MotivoDesalojo: "INTERRUPCION POR IO",
			IoType:         "DialFS",
			Interface:      words[1],
			TimeIO:         0,
		}

		sendFSDataToKernel(fileName, kind, 0, []int{0}, 0)

	case "IO_FS_DELETE":
		fileName := words[2]
		GLOBALrequestCPU = KernelRequest{
			PcbUpdated:     *contextoEjecucion,
			MotivoDesalojo: "INTERRUPCION POR IO",
			IoType:         "DialFS",
			Interface:      words[1],
			TimeIO:         0,
		}
		sendFSDataToKernel(fileName, kind, 0, []int{0}, 0)

	case "IO_FS_TRUNCATE":
		fileName := words[2]
		regTamano := words[3]
		valueLength := verificarRegistro(regTamano, contextoEjecucion)
		GLOBALrequestCPU = KernelRequest{
			PcbUpdated:     *contextoEjecucion,
			MotivoDesalojo: "INTERRUPCION POR IO",
			IoType:         "DialFS",
			Interface:      words[1],
			TimeIO:         0,
		}
		sendFSDataToKernel(fileName, kind, valueLength, []int{0}, 0)

	case "IO_FS_WRITE":
		fileName := words[2]

		regDirec := words[3]
		valueAdress := verificarRegistro(regDirec, contextoEjecucion)

		regTamano := words[4]
		valueLength := verificarRegistro(regTamano, contextoEjecucion)

		regPuntero := words[5]
		valuePuntero := verificarRegistro(regPuntero, contextoEjecucion)

		direcFisica := TranslateAddress(contextoEjecucion.Pid, valueAdress, GLOBALpageTam, valueLength)

		GLOBALrequestCPU = KernelRequest{
			PcbUpdated:     *contextoEjecucion,
			MotivoDesalojo: "INTERRUPCION POR IO",
			IoType:         "DialFS",
			Interface:      words[1],
			TimeIO:         0,
		}
		sendFSDataToKernel(fileName, kind, valueLength, direcFisica, valuePuntero)

	case "IO_FS_READ":
		fileName := words[2]

		regDirec := words[3]
		valueAdress := verificarRegistro(regDirec, contextoEjecucion)

		regTamano := words[4]
		valueLength := verificarRegistro(regTamano, contextoEjecucion)

		regPuntero := words[5]
		valuePuntero := verificarRegistro(regPuntero, contextoEjecucion)

		direcFisica := TranslateAddress(contextoEjecucion.Pid, valueAdress, GLOBALpageTam, valueLength)
		GLOBALrequestCPU = KernelRequest{
			PcbUpdated:     *contextoEjecucion,
			MotivoDesalojo: "INTERRUPCION POR IO",
			IoType:         "DialFS",
			Interface:      words[1],
			TimeIO:         0,
		}
		sendFSDataToKernel(fileName, kind, valueLength, direcFisica, valuePuntero)

	default:
		return fmt.Errorf("tipo de instrucción no soportado")
	}
	return nil
}

func verificarRegistro(registerName string, contextoEjecucion *PCB) int {
	var registerValue int
	switch registerName {
	case "AX":
		registerValue = int(contextoEjecucion.CpuReg.AX)
	case "BX":
		registerValue = int(contextoEjecucion.CpuReg.BX)
	case "CX":
		registerValue = int(contextoEjecucion.CpuReg.CX)
	case "DX":
		registerValue = int(contextoEjecucion.CpuReg.DX)
	case "SI":
		registerValue = int(contextoEjecucion.CpuReg.SI)
	case "DI":
		registerValue = int(contextoEjecucion.CpuReg.DI)
	case "EAX":
		registerValue = int(contextoEjecucion.CpuReg.EAX)
	case "EBX":
		registerValue = int(contextoEjecucion.CpuReg.EBX)
	case "ECX":
		registerValue = int(contextoEjecucion.CpuReg.ECX)
	case "EDX":
		registerValue = int(contextoEjecucion.CpuReg.EDX)
	default:
		log.Fatalf("Register %s not found", registerName)
	}
	return registerValue
}

func CheckSignal(w http.ResponseWriter, r *http.Request, pid int, motivo string, recurso string) error {
	waitRequest := ResponseWait{
		Recurso: recurso,
		Pid:     pid,
	}

	waitRequestJSON, err := json.Marshal(waitRequest)
	if err != nil {
		http.Error(w, "Error al codificar los datos JSON", http.StatusInternalServerError)
		return err
	}

	kernelURL := fmt.Sprintf("http://%s:%d/signal", globals.ClientConfig.IpKernel, globals.ClientConfig.PortKernel)
	resp, err := http.Post(kernelURL, "application/json", bytes.NewBuffer(waitRequestJSON))
	if err != nil {
		http.Error(w, "Error al enviar la solicitud al kernel", http.StatusInternalServerError)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Error en la respuesta del kernel", http.StatusInternalServerError)
		return err
	}

	var signalResponse struct {
		Success string `json:"success"`
	}

	err = json.NewDecoder(resp.Body).Decode(&signalResponse)
	if err != nil {
		http.Error(w, "Error al decodificar los datos JSON de la respuesta del kernel", http.StatusInternalServerError)
		return err
	}
	if signalResponse.Success == "exit" {
		err := TerminarProceso(&GLOBALcontextoDeEjecucion.CpuReg, "INVALID_RESOURCE")
		if err != nil {
			return fmt.Errorf("error en execute: %s", err)
		}
	}
	return nil
}

func CheckWait(w http.ResponseWriter, r *http.Request, registerCPU *PCB, recurso string) error {
	waitRequest := ResponseWait{
		Recurso: recurso,
		Pid:     registerCPU.Pid,
	}

	waitRequestJSON, err := json.Marshal(waitRequest)
	if err != nil {
		http.Error(w, "Error al codificar los datos JSON", http.StatusInternalServerError)
		return err
	}

	kernelURL := fmt.Sprintf("http://%s:%d/wait", globals.ClientConfig.IpKernel, globals.ClientConfig.PortKernel)
	resp, err := http.Post(kernelURL, "application/json", bytes.NewBuffer(waitRequestJSON))
	if err != nil {
		http.Error(w, "Error al enviar la solicitud al kernel", http.StatusInternalServerError)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Error en la respuesta del kernel", http.StatusInternalServerError)
		return err
	}

	var waitResponse struct {
		Success string `json:"success"`
	}

	err = json.NewDecoder(resp.Body).Decode(&waitResponse)
	if err != nil {
		http.Error(w, "Error al decodificar los datos JSON de la respuesta del kernel", http.StatusInternalServerError)
		return err
	}
	if waitResponse.Success == "false" {
		interrupt = true
		GLOBALrequestCPU = KernelRequest{
			MotivoDesalojo: "WAIT",
			Recurso:        recurso,
		}
	} else if waitResponse.Success == "exit" {
		err := TerminarProceso(&GLOBALcontextoDeEjecucion.CpuReg, "INVALID_RESOURCE")
		if err != nil {
			return fmt.Errorf("error en execute: %s", err)
		}
	}

	return nil
}

func Checkinterrupts(w http.ResponseWriter, r *http.Request) { // A chequear

	var responseInterruptLocal ResponseInterrupt

	err := json.NewDecoder(r.Body).Decode(&responseInterruptLocal)
	if err != nil {
		http.Error(w, "Error al decodificar los datos JSON", http.StatusInternalServerError)
		return
	}

	if responseInterruptLocal.Motivo == "INTERRUPTED_BY_USER" {
		// Siempre procesar INTERRUPTED_BY_USER inmediatamente
		responseInterruptGlobal = responseInterruptLocal
	} else if responseInterruptLocal.Motivo == "CLOCK" {
		// Verificar si ya hay una interrupción pendiente
		if responseInterruptGlobal.Interrupt && responseInterruptGlobal.Motivo == "INTERRUPTED_BY_USER" {
			// Ya hay una interrupción de mayor prioridad, ignorar CLOCK
		} else {
			// No hay interrupción de mayor prioridad, procesar CLOCK
			responseInterruptGlobal = responseInterruptLocal
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(interrupt)
}

func CheckTLB(pid, page int) (int, bool) { //Verifica si la etrada ya estaba en la globalTLB. Si se usa LRU, actualiza el tiempo de acceso
	for i, entry := range globalTLB {
		if entry.PID == pid && entry.Pagina == page {
			if replacementAlgorithm == "LRU" {
				globalTLB[i].UltimoAcceso = time.Now()
			}
			return entry.Frame, true
		}
	}
	return -1, false
}

func ReplaceTLBEntry(pid, page, frame int) { //Reemplazo una entrada de globalTLB según el algoritmo de reemplazo
	newEntry := TLBEntry{
		PID:                pid,
		Pagina:             page,
		Frame:              frame,
		UltimoAcceso:       time.Now(),
		globalPosicionFila: globalPosicionFila,
	}

	if len(globalTLB) < globalTLBsize {
		globalTLB = append(globalTLB, newEntry) //Si la globalTLB no está llena, agrego la entrada
	} else {
		if replacementAlgorithm == "FIFO" {
			victima := 0
			for i, entry := range globalTLB {
				if entry.globalPosicionFila < globalTLB[victima].globalPosicionFila {
					victima = i
				}
			}
			globalTLB[victima] = newEntry
		} else if replacementAlgorithm == "LRU" {
			victima := 0
			for i, entry := range globalTLB {
				if entry.UltimoAcceso.Before(globalTLB[victima].UltimoAcceso) {
					victima = i
				}
			}
			globalTLB[victima] = newEntry
		}
	}
	globalPosicionFila++
}

func TranslateHandler(w http.ResponseWriter, r *http.Request) {
	var req TranslationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Realizar la traducción
	addresses := TranslateAddress(req.PID, req.DireccionLogica, req.TamPag, req.TamData)

	// Responder con las direcciones físicas
	res := TranslationResponse{DireccionesFisicas: addresses}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func TranslateAddress(pid, DireccionLogica, TamPag, TamData int) []int {
	var DireccionesFisicas []int
	cache := make(map[int]int)    // Mapa para no buscar 10 veces el mismo marco
	tlbHits := make(map[int]bool) // Mapa para registrar los TLB hits por página

	for i := 0; i < TamData; i++ {
		pageNumber := int(math.Floor(float64(DireccionLogica) / float64(TamPag)))
		pageOffset := DireccionLogica - (pageNumber * TamPag)
		frame, found := CheckTLB(pid, pageNumber)

		if !found {
			if cachedFrame, ok := cache[pageNumber]; ok {
				frame = cachedFrame
			} else {
				log.Printf("PID: %d - TLB MISS - Página: %d", pid, pageNumber)
				err := FetchFrameFromMemory(pid, pageNumber)
				if err != nil {
					fmt.Println("Error al obtener el marco desde la memoria")
					return nil
				}
				frame = MemoryFrame
				cache[pageNumber] = frame
				log.Printf("PID: %d - OBTENER MARCO - Página: %d - Marco: %d", pid, pageNumber, frame)
				if globals.ClientConfig.NumberFellingTLB > 0 {
					ReplaceTLBEntry(pid, pageNumber, MemoryFrame)
				}
			}
		} else {
			// Verificar si ya se ha registrado un hit para esta página
			if !tlbHits[pageNumber] {
				log.Printf("PID: %d - TLB HIT - Página: %d", pid, pageNumber)
				tlbHits[pageNumber] = true
			}
		}

		DireccionFisica := frame*TamPag + pageOffset
		DireccionesFisicas = append(DireccionesFisicas, DireccionFisica)
		DireccionLogica++
	}
	return DireccionesFisicas
}

func TamRestantePagina(dirLog, tamPag int) int {
	// Calcular el offset dentro de la página actual
	offsetEnPagina := dirLog % tamPag

	// Calcular el tamaño restante en la página actual
	tamRestante := tamPag - offsetEnPagina

	return tamRestante
}

// simulacion de la obtención de un marco desde la memoria
func FetchFrameFromMemory(pid, pageNumber int) error {
	memoryURL := fmt.Sprintf("http://%s:%d/getFramefromCPU", globals.ClientConfig.IPMemory, globals.ClientConfig.PortMemory)
	var pageTable bodyPageTable
	pageTable.Pid = pid
	pageTable.Page = pageNumber

	pageTableJSON, err := json.Marshal(pageTable)
	if err != nil {
		log.Fatalf("Error al serializar el Input: %v", err)
	}

	resp, err := http.Post(memoryURL, "application/json", bytes.NewBuffer(pageTableJSON))
	if err != nil {
		log.Fatalf("error al enviar la solicitud al módulo de memoria: %v", err)
	}
	defer resp.Body.Close()
	return nil
}

func RecieveFramefromMemory(w http.ResponseWriter, r *http.Request) {
	var bodyFrame BodyFrame
	err := json.NewDecoder(r.Body).Decode(&bodyFrame)
	if err != nil {
		http.Error(w, "Error al decodificar los datos JSON", http.StatusInternalServerError)
		return
	}
	MemoryFrame = bodyFrame.Frame

	w.WriteHeader(http.StatusOK)
}

func sendResizeMemory(tam int) {
	memoriaURL := fmt.Sprintf("http://%s:%d/resizeProcess", globals.ClientConfig.IPMemory, globals.ClientConfig.PortMemory)
	var process bodyProcess
	process.Pid = GLOBALcontextoDeEjecucion.Pid
	process.Pages = tam

	bodyResizeJSON, err := json.Marshal(process)
	if err != nil {
		log.Fatalf("Error al serializar el Input: %v", err)
	}

	resp, err := http.Post(memoriaURL, "application/json", bytes.NewBuffer(bodyResizeJSON))
	if err != nil {
		log.Fatalf("error al enviar la solicitud al módulo de memoria: %v", err)
	}
	defer resp.Body.Close()

}

func sendREGtoKernel(adress []int, length int, pid int) {
	kernelURL := fmt.Sprintf("http://%s:%d/recieveREG", globals.ClientConfig.IpKernel, globals.ClientConfig.PortKernel)
	var BodyRegisters bodyRegisters
	BodyRegisters.Pid = pid
	BodyRegisters.DirFisica = adress
	BodyRegisters.LengthREG = length

	BodyRegistersJSON, err := json.Marshal(BodyRegisters)
	if err != nil {
		log.Fatalf("Error al serializar el Input: %v", err)
	}

	resp, err := http.Post(kernelURL, "application/json", bytes.NewBuffer(BodyRegistersJSON))
	if err != nil {
		log.Fatalf("error al enviar la solicitud al módulo de memoria: %v", err)
	}
	defer resp.Body.Close()
}

func sendFSDataToKernel(fileName string, instructionFS string, regTamano int, regDireccion []int, regPuntero int) {
	fsStructure := FSstructure{
		FileName:      fileName,
		FSInstruction: instructionFS,
		FSRegTam:      regTamano,
		FSRegDirec:    regDireccion,
		FSRegPuntero:  regPuntero,
	}
	kernelURL := fmt.Sprintf("http://%s:%d/recieveFSDATA", globals.ClientConfig.IpKernel, globals.ClientConfig.PortKernel)

	fsStructureJSON, err := json.Marshal(fsStructure)
	if err != nil {
		log.Fatalf("Error al serializar el Input: %v", err)
	}

	resp, err := http.Post(kernelURL, "application/json", bytes.NewBuffer(fsStructureJSON))
	if err != nil {
		log.Fatalf("error al enviar la solicitud al módulo de memoria: %v", err)
	}
	defer resp.Body.Close()
}

func ReceiveTamPage(w http.ResponseWriter, r *http.Request) {
	var req BodyPageTam
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	GLOBALpageTam = req.PageTam
	w.WriteHeader(http.StatusOK)
}
