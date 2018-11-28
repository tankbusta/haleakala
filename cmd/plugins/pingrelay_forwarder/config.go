package pingrelay_forwarder

type AuthorizedKeys struct {
	Name       string   `yaml:"name" json:"name"`
	PrivateKey string   `yaml:"server_private_key" json:"server_private_key"`
	PublicKey  string   `yaml:"client_public_key" json:"client_public_key"`
	Addressess []string `yaml:"addresses" json:"addresses"`
}

type Config struct {
	ClientAddress  string           `yaml:"client_address" json:"client_address"`
	ListenAddress  string           `yaml:"listen_address" json:"listen_address"`
	ServerDomain   string           `json:"server_domain"`
	ServerKey      string           `json:"server_key"`
	AuthorizedKeys []AuthorizedKeys `yaml:"authorized_clients" json:"authorized_clients"`
}
