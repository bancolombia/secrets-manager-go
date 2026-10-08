package envsm

import (
	"fmt"
	"os"
	"strings"

	"github.com/bancolombia/secrets-manager-go/api"
)

type EnvSecretsManager struct {
	settings  api.Settings
	lookupEnv func(string) (string, bool)
}

func NewEnvSecretsManager(settings api.Settings) *EnvSecretsManager {
	return &EnvSecretsManager{settings: settings, lookupEnv: os.LookupEnv}
}

// For testing: allow injecting a custom environment lookup
func NewEnvSecretsManagerWithLookup(settings api.Settings, lookupEnv func(string) (string, bool)) *EnvSecretsManager {
	return &EnvSecretsManager{settings: settings, lookupEnv: lookupEnv}
}

func (d *EnvSecretsManager) GetSecret(name string) (string, error) {
	if value, ok := d.lookupEnv(name); ok {
		return value, nil
	}
	normalized := normalizeEnvName(name)
	if normalized != name {
		if value, ok := d.lookupEnv(normalized); ok {
			return value, nil
		}
		return "", fmt.Errorf("secret %q not found as environment variable %q or %q", name, name, normalized)
	}
	return "", fmt.Errorf("secret %q not found as environment variable", name)
}

// normalizeEnvName converts a secret key to the environment variable name
// produced by tools like Kubernetes envFrom (upper case, '-' and '.' replaced
// by '_').
func normalizeEnvName(name string) string {
	return strings.NewReplacer("-", "_", ".", "_").Replace(strings.ToUpper(name))
}
