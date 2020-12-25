package config

// Config is the root config
type Config struct {
	Server *Server `json:"server,omitempty" yaml:"server,omitempty"`
}

// Server stores all server related configuration
type Server struct {
	Tls         *Tls   `json:"tls" yaml:"tls"`
	TcpAddress  string `json:"tcp,omitempty" yaml:"tcp,omitempty"`
	HttpAddress string `json:"http,omitempty" yaml:"http,omitempty"`
	MaxQos      byte `json:"max_qos,omitempty" yaml:"max_qos,omitempty"`
}

// Tls stores the TLS config for the server
type Tls struct {
	CertFile string `json:"cert,omitempty" yaml:"cert,omitempty"`
	KeyFile  string `json:"key,omitempty" yaml:"key,omitempty"`
}
