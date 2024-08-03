package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/kernel/utils"
)

func main() {
	utils.ConfigurarLogger()
	globals.ClientConfig = utils.IniciarConfiguracion(os.Args[1])

	if globals.ClientConfig == nil {
		log.Fatalf("No se pudo cargar la configuración")
	}

	puerto := globals.ClientConfig.Puerto

	http.HandleFunc("PUT /process", utils.IniciarProceso)

	http.HandleFunc("POST /syscall", utils.ProcessSyscall)
	http.HandleFunc("POST /SendPortOfInterfaceToKernel", utils.RecievePortOfInterfaceFromIO)
	http.HandleFunc("POST /recieveREG", utils.RecieveREGFromCPU)

	http.HandleFunc("POST /recieveFSDATA", utils.RecieveFileNameFromCPU)

	http.HandleFunc("DELETE /process", utils.FinalizarProceso)
	http.HandleFunc("POST /wait", utils.RecieveWait)
	http.HandleFunc("POST /signal", utils.HandleSignal)
	http.HandleFunc("GET /process/{pid}", utils.EstadoProceso)
	http.HandleFunc("PUT /plani", utils.IniciarPlanificacion)
	http.HandleFunc("DELETE /plani", utils.DetenerPlanificacion)
	http.HandleFunc("GET /process", utils.ListarProcesos)
	http.ListenAndServe(":"+strconv.Itoa(puerto), nil)
}
