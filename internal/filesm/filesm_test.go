package filesm

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/bancolombia/secrets-manager-go/api"
)

func settingsWithPath(path interface{}) api.Settings {
	return api.Settings{VaultType: "file", VaultConfig: map[string]interface{}{"path": path}}
}

func writeSecretFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatalf("unable to prepare test file: %v", err)
	}
}

func TestGetSecret_Success(t *testing.T) {
	dir := t.TempDir()
	writeSecretFile(t, dir, "my-secret", "value\n")
	mgr := NewFileSecretsManager(settingsWithPath(dir))
	secret, err := mgr.GetSecret("my-secret")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if secret != "value" {
		t.Errorf("expected 'value', got '%s'", secret)
	}
}

func TestGetSecret_TrimsTrailingCRLF(t *testing.T) {
	dir := t.TempDir()
	writeSecretFile(t, dir, "my-secret", "value\r\n")
	mgr := NewFileSecretsManager(settingsWithPath(dir))
	secret, err := mgr.GetSecret("my-secret")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if secret != "value" {
		t.Errorf("expected 'value', got '%q'", secret)
	}
}

func TestGetSecret_PreservesWhitespace(t *testing.T) {
	dir := t.TempDir()
	writeSecretFile(t, dir, "my-secret", "  value \t")
	mgr := NewFileSecretsManager(settingsWithPath(dir))
	secret, err := mgr.GetSecret("my-secret")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if secret != "  value \t" {
		t.Errorf("expected whitespace to be preserved, got '%q'", secret)
	}
}

func TestGetSecret_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	writeSecretFile(t, dir, "my-secret", "")
	mgr := NewFileSecretsManager(settingsWithPath(dir))
	secret, err := mgr.GetSecret("my-secret")
	if err != nil {
		t.Fatalf("expected no error for empty file, got %v", err)
	}
	if secret != "" {
		t.Errorf("expected empty secret, got '%s'", secret)
	}
}

func TestGetSecret_MissingFile(t *testing.T) {
	mgr := NewFileSecretsManager(settingsWithPath(t.TempDir()))
	_, err := mgr.GetSecret("absent-secret")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("expected not-exist error, got %v", err)
	}
}

func TestGetSecret_ReadError(t *testing.T) {
	readErr := errors.New("permission denied")
	mgr := NewFileSecretsManagerWithReader(settingsWithPath("/mnt/test"), func(string) ([]byte, error) {
		return nil, readErr
	})
	_, err := mgr.GetSecret("my-secret")
	if !errors.Is(err, readErr) {
		t.Fatalf("expected wrapped read error, got %v", err)
	}
}

func TestGetSecret_DefaultPath(t *testing.T) {
	var gotPath string
	mgr := NewFileSecretsManagerWithReader(api.Settings{VaultType: "file"}, func(path string) ([]byte, error) {
		gotPath = path
		return []byte("value"), nil
	})
	secret, err := mgr.GetSecret("my-secret")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if want := filepath.Join(defaultPath, "my-secret"); gotPath != want {
		t.Errorf("expected path '%s', got '%s'", want, gotPath)
	}
	if secret != "value" {
		t.Errorf("expected 'value', got '%s'", secret)
	}
}

func TestGetSecret_PathNotAString(t *testing.T) {
	mgr := NewFileSecretsManager(settingsWithPath(123))
	_, err := mgr.GetSecret("my-secret")
	want := `vault config "path" must be a string, got int`
	if err == nil || err.Error() != want {
		t.Fatalf("expected error '%s', got %v", want, err)
	}
}

func TestGetSecret_EmptyPath(t *testing.T) {
	mgr := NewFileSecretsManager(settingsWithPath(""))
	_, err := mgr.GetSecret("my-secret")
	want := `vault config "path" must not be empty`
	if err == nil || err.Error() != want {
		t.Fatalf("expected error '%s', got %v", want, err)
	}
}

func TestGetSecret_BlankPath(t *testing.T) {
	mgr := NewFileSecretsManager(settingsWithPath("   "))
	_, err := mgr.GetSecret("my-secret")
	want := `vault config "path" must not be empty`
	if err == nil || err.Error() != want {
		t.Fatalf("expected error '%s', got %v", want, err)
	}
}

func TestGetSecret_RejectsEmptyName(t *testing.T) {
	mgr := NewFileSecretsManagerWithReader(settingsWithPath("/mnt/test"), func(string) ([]byte, error) {
		t.Error("reader must not be called for an invalid name")
		return nil, nil
	})
	_, err := mgr.GetSecret("")
	want := "invalid secret name: must not be empty"
	if err == nil || err.Error() != want {
		t.Fatalf("expected error '%s', got %v", want, err)
	}
}

func TestGetSecret_RejectsPathSeparator(t *testing.T) {
	mgr := NewFileSecretsManagerWithReader(settingsWithPath("/mnt/test"), func(string) ([]byte, error) {
		t.Error("reader must not be called for an invalid name")
		return nil, nil
	})
	_, err := mgr.GetSecret("a/b")
	want := `invalid secret name "a/b": must not contain path separators`
	if err == nil || err.Error() != want {
		t.Fatalf("expected error '%s', got %v", want, err)
	}
}

func TestGetSecret_RejectsBackslashSeparator(t *testing.T) {
	mgr := NewFileSecretsManagerWithReader(settingsWithPath("/mnt/test"), func(string) ([]byte, error) {
		t.Error("reader must not be called for an invalid name")
		return nil, nil
	})
	_, err := mgr.GetSecret(`a\b`)
	if err == nil || err.Error() != `invalid secret name "a\\b": must not contain path separators` {
		t.Fatalf("expected path separator error, got %v", err)
	}
}

func TestGetSecret_RejectsBareParentDir(t *testing.T) {
	mgr := NewFileSecretsManagerWithReader(settingsWithPath("/mnt/test"), func(string) ([]byte, error) {
		t.Error("reader must not be called for an invalid name")
		return nil, nil
	})
	_, err := mgr.GetSecret("..")
	want := `invalid secret name "..": must not contain ".."`
	if err == nil || err.Error() != want {
		t.Fatalf("expected error '%s', got %v", want, err)
	}
}

func TestGetSecret_RejectsDotDotInsideName(t *testing.T) {
	mgr := NewFileSecretsManagerWithReader(settingsWithPath("/mnt/test"), func(string) ([]byte, error) {
		t.Error("reader must not be called for an invalid name")
		return nil, nil
	})
	_, err := mgr.GetSecret("a..b")
	want := `invalid secret name "a..b": must not contain ".."`
	if err == nil || err.Error() != want {
		t.Fatalf("expected error '%s', got %v", want, err)
	}
}

func TestGetSecret_RejectsTraversalAttempt(t *testing.T) {
	mgr := NewFileSecretsManagerWithReader(settingsWithPath("/mnt/test"), func(string) ([]byte, error) {
		t.Error("reader must not be called for an invalid name")
		return nil, nil
	})
	_, err := mgr.GetSecret("../../etc/passwd")
	if err == nil {
		t.Fatal("expected error for traversal attempt, got nil")
	}
}

func TestGetSecret_FollowsSymlink(t *testing.T) {
	dir := t.TempDir()
	dataDir := filepath.Join(dir, "..data")
	if err := os.Mkdir(dataDir, 0o700); err != nil {
		t.Fatalf("unable to prepare data dir: %v", err)
	}
	writeSecretFile(t, dataDir, "my-secret", "symlinked\n")
	if err := os.Symlink(filepath.Join("..data", "my-secret"), filepath.Join(dir, "my-secret")); err != nil {
		t.Fatalf("unable to create symlink: %v", err)
	}
	mgr := NewFileSecretsManager(settingsWithPath(dir))
	secret, err := mgr.GetSecret("my-secret")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if secret != "symlinked" {
		t.Errorf("expected 'symlinked', got '%s'", secret)
	}
}

func TestNewFileSecretsManagerWithReader(t *testing.T) {
	mgr := NewFileSecretsManagerWithReader(settingsWithPath("/mnt/test"), func(string) ([]byte, error) {
		return []byte("injected"), nil
	})
	if mgr.path != "/mnt/test" {
		t.Errorf("expected path '/mnt/test', got '%s'", mgr.path)
	}
	if mgr.configErr != nil {
		t.Errorf("expected no config error, got %v", mgr.configErr)
	}
	secret, err := mgr.GetSecret("my-secret")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if secret != "injected" {
		t.Errorf("expected 'injected', got '%s'", secret)
	}
}
