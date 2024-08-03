package globals

type Config struct {
	Puerto           int    `json:"port"`
	MemorySize       int    `json:"memory_size"`
	IpCPU            string `json:"ip_cpu"`
	IpKernel         string `json:"ip_kernel"`
	IpEntradaSalida  string `json:"ip_entradasalida"`
	PuertoCPU        int    `json:"port_cpu"`
	PuertoKernel     int    `json:"port_kernel"`
	PageSize         int    `json:"page_size"`
	InstructionsPath string `json:"instructions_path"`
	DelayResponse    int    `json:"delay_response"`
}

var ClientConfig *Config
