package secretsmanager

import (
	"errors"
	"testing"

	"github.com/bancolombia/secretsmanager/api"
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
