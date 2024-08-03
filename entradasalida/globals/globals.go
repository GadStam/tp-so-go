package globals

type Config struct {
	Puerto                int    `json:"port"`
	Tipo                  string `json:"type"`
	UnidadDeTiempo        int    `json:"unit_work_time"`
	IPKernel              string `json:"ip_kernel"`
	PuertoKernel          int    `json:"port_kernel"`
	IPMemoria             string `json:"ip_memory"`
	PuertoMemoria         int    `json:"port_memory"`
	PathDialFS            string `json:"dialfs_path"`
	TamanioBloqueDialFS   int    `json:"dialfs_block_size"`
	CantidadBloquesDialFS int    `json:"dialfs_block_count"`
}

var ClientConfig *Config

type Interfaces struct {
	Nombre string
	Config *Config
}
