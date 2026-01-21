package awssm

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/bancolombia/secretsmanager/api"
)

type SecretsManagerClient interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

type AwsSecretsManager struct {
	settings api.Settings
	config   aws.Config
	client   SecretsManagerClient
}

func NewAwsSecretsManager(settings api.Settings) *AwsSecretsManager {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(getRegionFromConfig(settings)))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	client := secretsmanager.NewFromConfig(cfg)
	return &AwsSecretsManager{settings: settings, config: cfg, client: client}
}

// For testing: allow injecting a custom client
func NewAwsSecretsManagerWithClient(settings api.Settings, client SecretsManagerClient) *AwsSecretsManager {
	cfg, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion(getRegionFromConfig(settings)))
	return &AwsSecretsManager{settings: settings, config: cfg, client: client}
}

func (d *AwsSecretsManager) GetSecret(name string) (string, error) {
	out, err := d.client.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(name),
	})
	if err != nil {
		log.Printf("unable to get secret %v, %v", name, err)
		return "", err
	}
	return aws.ToString(out.SecretString), nil
}

func getRegionFromConfig(settings api.Settings) string {
	m := settings.VaultConfig
	if region, ok := m["region"]; ok {
		return region.(string)
	}
	return "us-east-1"
}
