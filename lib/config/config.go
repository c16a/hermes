package config

// Config is the root config
type Config struct {
	Server *Server `json:"server,omitempty" yaml:"server,omitempty"`
}

// Server stores all server related configuration
type Server struct {
	Tls         *Tls     `json:"tls" yaml:"tls"`
	TcpAddress  string   `json:"tcp,omitempty" yaml:"tcp,omitempty"`
	HttpAddress string   `json:"http,omitempty" yaml:"http,omitempty"`
	MaxQos      byte     `json:"max_qos,omitempty" yaml:"max_qos,omitempty"`
	Auth        *Auth    `json:"auth,omitempty" yaml:"auth,omitempty"`
	Offline     *Offline `json:"offline,omitempty"`
}

// Tls stores the TLS config for the server
type Tls struct {
	CertFile string `json:"cert,omitempty" yaml:"cert,omitempty"`
	KeyFile  string `json:"key,omitempty" yaml:"key,omitempty"`
}

type Auth struct {
	Type     string `json:"type,omitempty" yaml:"type,omitempty"`
	LdapHost string `json:"ldap_host,omitempty" yaml:"ldap_host,omitempty"`
	LdapPort int    `json:"ldap_port,omitempty" yaml:"ldap_port,omitempty"`
	LdapDn   string `json:"ldap_dn,omitempty" yaml:"ldap_dn,omitempty"`
}

type Offline struct {
	Path         string `json:"path,omitempty"`
	MaxTableSize int64  `json:"max_table_size,omitempty"`
	NumTables    int    `json:"num_tables,omitempty"`
}
