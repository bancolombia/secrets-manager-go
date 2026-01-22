package api

type SecretReader interface {
	GetSecret(name string) (string, error)
}

type Settings struct {
	VaultType   string
	VaultConfig map[string]interface{}
}
