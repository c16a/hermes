package config

// Config is the root config
type Config struct {
	Server *Server `json:"server" yaml:"server"`
}

// Server stores all server related configuration
type Server struct {
	Tls         *Tls   `json:"tls" yaml:"tls"`
	TcpAddress  string `json:"tcp" yaml:"tcp"`
	HttpAddress string `json:"http" yaml:"http"`
}

// Tls stores the TLS config for the server
type Tls struct {
	CertFile string `json:"cert" yaml:"cert"`
	KeyFile  string `json:"key" yaml:"key"`
}
