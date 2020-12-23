package config

type Config struct {
	Server *Server `json:"server" yaml:"server"`
}

type Server struct {
	Tls         *Tls   `json:"tls" yaml:"tls"`
	TcpAddress  string `json:"tcp" yaml:"tcp"`
	HttpAddress string `json:"http" yaml:"http"`
}

type Tls struct {
	CertFile string `json:"cert" yaml:"cert"`
	KeyFile  string `json:"key" yaml:"key"`
}
