package common

type SSHConfig struct {
	PrivateKey     string `json:"privateKey"`
	PublicKey      string `json:"publicKey"`
	AuthorizedKeys string `json:"authorizedKeys"`
	Config         string `json:"config"`
}
