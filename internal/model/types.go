package model

// PortMapping représente une correspondance de port host:container.
type PortMapping struct {
	Host      string `json:"host"`
	Container string `json:"container"`
	Protocol  string `json:"protocol,omitempty"` // tcp | udp
}

// VolumeMount représente le montage d'un volume dans un service.
type VolumeMount struct {
	Source string `json:"source"`
	Target string `json:"target"`
}
