package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	//"github.com/sisoputnfrba/tp-golang/entradasalida/globals"
	"github.com/sisoputnfrba/tp-golang/entradasalida/utils"
)

func main() {
	interfaceName := os.Args[1]
	pathToConfig := os.Args[2]

	config, err := utils.LoadConfig(pathToConfig)
	if err != nil {
		log.Fatalf("Error al cargar la configuración desde '%s': %v", pathToConfig, err)
	}
	utils.ConfigurarLogger(interfaceName, config)
	utils.SendPortOfInterfaceToKernel(interfaceName, config)
	Puerto := config.Puerto

	//http.HandleFunc("GET /input", utils.Prueba)
	http.HandleFunc("POST /recieveREG", utils.RecieveREG)
	http.HandleFunc("POST /recieveFSDATA", utils.RecieveFSDataFromKernel)
	http.HandleFunc("/interfaz", utils.Iniciar)
	http.HandleFunc("/receiveContentFromMemory", utils.ReceiveContentFromMemory)

	// Cargar la configuración desde el archivo

	http.ListenAndServe(":"+strconv.Itoa(Puerto), nil)
}
