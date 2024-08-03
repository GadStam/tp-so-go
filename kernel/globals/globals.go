package globals

type Config struct {
	Puerto                 int      `json:"port"`
	IpMemoria              string   `json:"ip_memory"`
	PuertoMemoria          int      `json:"port_memory"`
	IpCPU                  string   `json:"ip_cpu"`
	IpEntradaSalida        string   `json:"ip_entradasalida"`
	PuertoCPU              int      `json:"port_cpu"`
	AlgoritmoPlanificacion string   `json:"planning_algorithm"`
	Quantum                int      `json:"quantum"`
	Recursos               []string `json:"resources"`
	InstanciasRecursos     []int    `json:"resource_instances"`
	Multiprogramacion      int      `json:"multiprogramming"`
}

var ClientConfig *Config
