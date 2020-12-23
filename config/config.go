package config

type Config struct {
	Server *Server `json:"server" yaml:"server"`
}

type Server struct {
	TcpAddress  string `json:"tcp" yaml:"tcp"`
	HttpAddress string `json:"http" yaml:"http"`
}
