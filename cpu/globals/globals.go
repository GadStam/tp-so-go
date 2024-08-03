package globals

type Config struct {
	Puerto           int    `json:"port"`
	IpKernel         string `json:"ip_kernel"`
	IPMemory         string `json:"ip_memory"`
	PortMemory       int    `json:"port_memory"`
	PortKernel       int    `json:"port_kernel"`
	NumberFellingTLB int    `json:"number_felling_tlb"`
	AlgorithmTLB     string `json:"algorithm_tlb"`
}

var ClientConfig *Config
