package filesm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bancolombia/secretsmanager/api"
)

const defaultPath = "/mnt/secrets-store"

type FileSecretsManager struct {
	settings  api.Settings
	path      string
	configErr error
	readFile  func(string) ([]byte, error)
}

func NewFileSecretsManager(settings api.Settings) *FileSecretsManager {
	path, configErr := getPathFromConfig(settings)
	return &FileSecretsManager{settings: settings, path: path, configErr: configErr, readFile: os.ReadFile}
}

// For testing: allow injecting a custom file reader
func NewFileSecretsManagerWithReader(settings api.Settings, readFile func(string) ([]byte, error)) *FileSecretsManager {
	path, configErr := getPathFromConfig(settings)
	return &FileSecretsManager{settings: settings, path: path, configErr: configErr, readFile: readFile}
}

func (d *FileSecretsManager) GetSecret(name string) (string, error) {
	if d.configErr != nil {
		return "", d.configErr
	}
	if err := validateSecretName(name); err != nil {
		return "", err
	}
	fullPath := filepath.Join(d.path, name)
	data, err := d.readFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("unable to read secret %q from %q: %w", name, fullPath, err)
	}
	return trimValue(string(data)), nil
}

func validateSecretName(name string) error {
	if name == "" {
		return fmt.Errorf("invalid secret name: must not be empty")
	}
	if strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("invalid secret name %q: must not contain path separators", name)
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("invalid secret name %q: must not contain %q", name, "..")
	}
	return nil
}

// trimValue removes trailing newline characters commonly appended when secrets
// are written to files (e.g. with echo); spaces and tabs are preserved.
func trimValue(value string) string {
	return strings.TrimRight(value, "\r\n")
}

func getPathFromConfig(settings api.Settings) (string, error) {
	if path, ok := settings.VaultConfig["path"]; ok {
		p, ok := path.(string)
		if !ok {
			return "", fmt.Errorf("vault config %q must be a string, got %T", "path", path)
		}
		if strings.TrimSpace(p) == "" {
			return "", fmt.Errorf("vault config %q must not be empty", "path")
		}
		return p, nil
	}
	return defaultPath, nil
}
