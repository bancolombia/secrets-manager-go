package envsm

import (
	"testing"

	"github.com/bancolombia/secretsmanager/api"
)

func TestGetSecret_Success(t *testing.T) {
	t.Setenv("MY_SECRET", "value")
	mgr := NewEnvSecretsManager(api.Settings{VaultType: "env"})
	secret, err := mgr.GetSecret("MY_SECRET")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if secret != "value" {
		t.Errorf("expected 'value', got '%s'", secret)
	}
}

func TestGetSecret_EmptyValueIsFound(t *testing.T) {
	t.Setenv("MY_EMPTY_SECRET", "")
	mgr := NewEnvSecretsManager(api.Settings{VaultType: "env"})
	secret, err := mgr.GetSecret("MY_EMPTY_SECRET")
	if err != nil {
		t.Fatalf("expected no error for set-but-empty variable, got %v", err)
	}
	if secret != "" {
		t.Errorf("expected empty secret, got '%s'", secret)
	}
}

func TestGetSecret_NormalizedFallbackWithDash(t *testing.T) {
	t.Setenv("MY_SECRET", "normalized")
	mgr := NewEnvSecretsManager(api.Settings{VaultType: "env"})
	secret, err := mgr.GetSecret("my-secret")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if secret != "normalized" {
		t.Errorf("expected 'normalized', got '%s'", secret)
	}
}

func TestGetSecret_NormalizedFallbackWithDot(t *testing.T) {
	t.Setenv("DB_PASSWORD", "dot-value")
	mgr := NewEnvSecretsManager(api.Settings{VaultType: "env"})
	secret, err := mgr.GetSecret("db.password")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if secret != "dot-value" {
		t.Errorf("expected 'dot-value', got '%s'", secret)
	}
}

func TestGetSecret_ExactNameWins(t *testing.T) {
	t.Setenv("my-secret", "exact")
	t.Setenv("MY_SECRET", "normalized")
	mgr := NewEnvSecretsManager(api.Settings{VaultType: "env"})
	secret, err := mgr.GetSecret("my-secret")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if secret != "exact" {
		t.Errorf("expected exact match 'exact' to win, got '%s'", secret)
	}
}

func TestGetSecret_NotFound(t *testing.T) {
	mgr := NewEnvSecretsManager(api.Settings{VaultType: "env"})
	_, err := mgr.GetSecret("MISSING_SECRET_FOR_TEST")
	want := `secret "MISSING_SECRET_FOR_TEST" not found as environment variable`
	if err == nil || err.Error() != want {
		t.Fatalf("expected error '%s', got %v", want, err)
	}
}

func TestGetSecret_NotFoundNormalizedTried(t *testing.T) {
	mgr := NewEnvSecretsManager(api.Settings{VaultType: "env"})
	_, err := mgr.GetSecret("missing-secret-for-test")
	want := `secret "missing-secret-for-test" not found as environment variable "missing-secret-for-test" or "MISSING_SECRET_FOR_TEST"`
	if err == nil || err.Error() != want {
		t.Fatalf("expected error '%s', got %v", want, err)
	}
}

func TestNewEnvSecretsManagerWithLookup(t *testing.T) {
	lookup := func(name string) (string, bool) {
		if name == "MY_SECRET" {
			return "injected", true
		}
		return "", false
	}
	mgr := NewEnvSecretsManagerWithLookup(api.Settings{VaultType: "env"}, lookup)
	secret, err := mgr.GetSecret("my-secret")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if secret != "injected" {
		t.Errorf("expected 'injected', got '%s'", secret)
	}
}
