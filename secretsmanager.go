package secretsmanager

import (
	"fmt"
	"log"
	"strings"

	"github.com/bancolombia/secretsmanager/api"
	"github.com/bancolombia/secretsmanager/internal/awssm"
)

type SecretsManager struct {
	Settings api.Settings
	vault    api.SecretReader
}

const VaultTypeAwsSecretManager = "awssm"

type noOpVault struct{}

func (d *noOpVault) GetSecret(name string) (string, error) {
	return "", fmt.Errorf("unsupported secret repository vault type")
}

func NewWithDefaults() *SecretsManager {
	// Default to AWS Secrets Manager in us-east-1
	m := make(map[string]interface{})
	m["region"] = "us-east-1"
	defaultSettings := &api.Settings{
		VaultType:   VaultTypeAwsSecretManager,
		VaultConfig: m,
	}
	return NewSecretsManager(*defaultSettings)
}

func NewSecretsManager(settings api.Settings) *SecretsManager {
	var vaultDef api.SecretReader
	switch strings.ToLower(settings.VaultType) {
	case VaultTypeAwsSecretManager:
		vaultDef = awssm.NewAwsSecretsManager(settings)
	default:
		log.Printf("unsupported backend [%s], using dummy backend instead", settings.VaultType)
		vaultDef = &noOpVault{}
	}
	return &SecretsManager{Settings: settings, vault: vaultDef}
}

func (s *SecretsManager) PullSecret(name string) (string, error) {
	out, err := s.vault.GetSecret(name)
	if err != nil {
		log.Printf("unable to get secret %v, %v", name, err)
		return "", err
	}
	return out, nil
}
