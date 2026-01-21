package api

import (
	"errors"
	"reflect"
	"testing"
)

type mockVault struct {
	secret string
	err    error
}

func (m *mockVault) GetSecret(name string) (string, error) {
	return m.secret, m.err
}

func TestGenericVault_GetSecret_Success(t *testing.T) {
	vault := &mockVault{secret: "mysecret", err: nil}
	secret, err := vault.GetSecret("key")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if secret != "mysecret" {
		t.Errorf("expected 'mysecret', got %v", secret)
	}
}

func TestGenericVault_GetSecret_Error(t *testing.T) {
	vault := &mockVault{secret: "", err: errors.New("fail")}
	_, err := vault.GetSecret("key")
	if err == nil || err.Error() != "fail" {
		t.Fatalf("expected error 'fail', got %v", err)
	}
}

func TestSettings_Struct(t *testing.T) {
	cfg := map[string]interface{}{"region": "us-east-1"}
	settings := Settings{VaultType: "aws", VaultConfig: cfg}
	if settings.VaultType != "aws" {
		t.Errorf("expected VaultType 'aws', got %v", settings.VaultType)
	}
	if !reflect.DeepEqual(settings.VaultConfig, cfg) {
		t.Errorf("expected VaultConfig %v, got %v", cfg, settings.VaultConfig)
	}
}
