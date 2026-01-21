package api

type GenericVault interface {
	GetSecret(name string) (string, error)
}

type Settings struct {
	VaultType   string
	VaultConfig map[string]interface{}
}
