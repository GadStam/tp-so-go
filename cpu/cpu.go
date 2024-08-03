package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	"github.com/sisoputnfrba/tp-golang/cpu/utils"
)

func main() {
	utils.ConfigurarLogger()
	globals.ClientConfig = utils.IniciarConfiguracion(os.Args[1])

	if globals.ClientConfig == nil {
		log.Fatalf("No se pudo cargar la configuración")
	}
	puerto := globals.ClientConfig.Puerto

	http.HandleFunc("/receivePCB", utils.ReceivePCB)
	http.HandleFunc("POST /receiveDataFromMemory", utils.RecieveMOV_IN)
	http.HandleFunc("/interrupt", utils.Checkinterrupts)
	http.HandleFunc("/translate", utils.TranslateHandler)
	http.HandleFunc("/recievePageTam", utils.ReceiveTamPage)
	http.HandleFunc("POST /recieveFrame", utils.RecieveFramefromMemory)
	http.ListenAndServe(":"+strconv.Itoa(puerto), nil)
}
