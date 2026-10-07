package secretsmanager

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/bancolombia/secrets-manager-go/api"
	"github.com/bancolombia/secrets-manager-go/internal/envsm"
	"github.com/bancolombia/secrets-manager-go/internal/filesm"
)

type mockVault struct {
	secret string
	err    error
}

func (m *mockVault) GetSecret(name string) (string, error) {
	return m.secret, m.err
}

func TestNewWithDefaults(t *testing.T) {
	mgr := NewWithDefaults()
	if mgr == nil {
		t.Fatal("expected non-nil SecretsManager")
	}
	if mgr.Settings.VaultType != VaultTypeAwsSecretManager {
		t.Errorf("expected VaultType %s, got %s", VaultTypeAwsSecretManager, mgr.Settings.VaultType)
	}
	region, ok := mgr.Settings.VaultConfig["region"]
	if !ok || region != "us-east-1" {
		t.Errorf("expected region 'us-east-1', got %v", region)
	}
}

func TestNewSecretsManager_AWS(t *testing.T) {
	settings := api.Settings{VaultType: VaultTypeAwsSecretManager}
	mgr := NewSecretsManager(settings)
	if mgr == nil {
		t.Fatal("expected non-nil SecretsManager")
	}
	if mgr.Settings.VaultType != VaultTypeAwsSecretManager {
		t.Errorf("expected VaultType %s, got %s", VaultTypeAwsSecretManager, mgr.Settings.VaultType)
	}
}

func TestNewSecretsManager_Unsupported(t *testing.T) {
	settings := api.Settings{VaultType: "unsupported"}
	mgr := NewSecretsManager(settings)
	if mgr == nil {
		t.Fatal("expected non-nil SecretsManager")
	}
	_, err := mgr.vault.GetSecret("any")
	if err == nil || err.Error() != "unsupported secret repository vault type" {
		t.Errorf("expected error for unsupported backend, got %v", err)
	}
}

func TestPullSecret_Success(t *testing.T) {
	mgr := &SecretsManager{
		Settings: api.Settings{VaultType: "mock"},
		vault:    &mockVault{secret: "value", err: nil},
	}
	secret, err := mgr.PullSecret("key")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if secret != "value" {
		t.Errorf("expected 'value', got %v", secret)
	}
}

func TestPullSecret_Error(t *testing.T) {
	mgr := &SecretsManager{
		Settings: api.Settings{VaultType: "mock"},
		vault:    &mockVault{secret: "", err: errors.New("fail")},
	}
	_, err := mgr.PullSecret("key")
	if err == nil || err.Error() != "fail" {
		t.Fatalf("expected error 'fail', got %v", err)
	}
}

func TestNoOpVault_GetSecret(t *testing.T) {
	vault := &noOpVault{}
	_, err := vault.GetSecret("any")
	if err == nil || err.Error() != "unsupported secret repository vault type" {
		t.Errorf("expected error for noOpVault, got %v", err)
	}
}

func TestVaultTypeConstants(t *testing.T) {
	if VaultTypeAwsSecretManager != "awssm" {
		t.Errorf("expected VaultTypeAwsSecretManager 'awssm', got '%s'", VaultTypeAwsSecretManager)
	}
	if VaultTypeEnv != "env" {
		t.Errorf("expected VaultTypeEnv 'env', got '%s'", VaultTypeEnv)
	}
	if VaultTypeFile != "file" {
		t.Errorf("expected VaultTypeFile 'file', got '%s'", VaultTypeFile)
	}
}

func TestNewSecretsManager_Env(t *testing.T) {
	mgr := NewSecretsManager(api.Settings{VaultType: VaultTypeEnv})
	if _, ok := mgr.vault.(*envsm.EnvSecretsManager); !ok {
		t.Errorf("expected *envsm.EnvSecretsManager, got %T", mgr.vault)
	}
}

func TestNewSecretsManager_EnvVaultTypeIsCaseInsensitive(t *testing.T) {
	mgr := NewSecretsManager(api.Settings{VaultType: "ENV"})
	if _, ok := mgr.vault.(*envsm.EnvSecretsManager); !ok {
		t.Errorf("expected *envsm.EnvSecretsManager, got %T", mgr.vault)
	}
}

func TestNewSecretsManager_File(t *testing.T) {
	settings := api.Settings{
		VaultType:   VaultTypeFile,
		VaultConfig: map[string]interface{}{"path": "/mnt/test"},
	}
	mgr := NewSecretsManager(settings)
	if _, ok := mgr.vault.(*filesm.FileSecretsManager); !ok {
		t.Errorf("expected *filesm.FileSecretsManager, got %T", mgr.vault)
	}
}

func TestPullSecret_EnvEndToEnd(t *testing.T) {
	t.Setenv("E2E_ENV_SECRET", "from-env")
	mgr := NewSecretsManager(api.Settings{VaultType: VaultTypeEnv})
	secret, err := mgr.PullSecret("e2e-env-secret")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if secret != "from-env" {
		t.Errorf("expected 'from-env', got '%s'", secret)
	}
}

func TestPullSecret_FileEndToEnd(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "e2e-file-secret"), []byte("from-file\n"), 0o600); err != nil {
		t.Fatalf("unable to prepare test file: %v", err)
	}
	settings := api.Settings{
		VaultType:   VaultTypeFile,
		VaultConfig: map[string]interface{}{"path": dir},
	}
	mgr := NewSecretsManager(settings)
	secret, err := mgr.PullSecret("e2e-file-secret")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if secret != "from-file" {
		t.Errorf("expected 'from-file', got '%s'", secret)
	}
}
