package awssm

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/bancolombia/secretsmanager/api"
)

type mockSecretsManagerClient struct {
	getSecretValueFunc func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

func (m *mockSecretsManagerClient) GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	return m.getSecretValueFunc(ctx, params, optFns...)
}

func TestGetSecret_Success(t *testing.T) {
	mockClient := &mockSecretsManagerClient{
		getSecretValueFunc: func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
			return &secretsmanager.GetSecretValueOutput{SecretString: params.SecretId}, nil
		},
	}
	m := make(map[string]interface{})
	m["region"] = "us-east-1"
	settings := api.Settings{VaultConfig: m, VaultType: "awssm"}

	mgr := NewAwsSecretsManagerWithClient(settings, mockClient)
	secret, err := mgr.GetSecret("my-secret")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if secret != "my-secret" {
		t.Errorf("expected secret 'my-secret', got '%s'", secret)
	}
}

func TestGetSecret_Error(t *testing.T) {
	mockClient := &mockSecretsManagerClient{
		getSecretValueFunc: func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
			return nil, errors.New("some error")
		},
	}
	m := make(map[string]interface{})
	m["region"] = "us-east-1"
	settings := api.Settings{VaultConfig: m, VaultType: "awssm"}
	mgr := NewAwsSecretsManagerWithClient(settings, mockClient)
	_, err := mgr.GetSecret("fail-secret")
	//defer func() {
	//	if r := recover(); r == nil {
	//		t.Errorf("expected log.Fatalf to exit, but it did not")
	//	}
	//}()
	if err.Error() != "some error" {
		t.Fatalf("expected error, got %v", err)
	}
}

func TestNewAwsSecretsManagerWithClient(t *testing.T) {
	mockClient := &mockSecretsManagerClient{
		getSecretValueFunc: func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
			return &secretsmanager.GetSecretValueOutput{SecretString: params.SecretId}, nil
		},
	}
	m := make(map[string]interface{})
	m["region"] = "us-east-1"
	settings := api.Settings{VaultConfig: m, VaultType: "awssm"}

	mgr := NewAwsSecretsManagerWithClient(settings, mockClient)
	if mgr.client != mockClient {
		t.Errorf("expected injected client to be used")
	}
	if mgr.settings.VaultConfig["region"] != "us-east-1" {
		t.Errorf("expected region to be 'us-east-1'")
	}
}
