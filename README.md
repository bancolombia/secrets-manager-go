# SecretsManager Go - Bancolombia (alpha)

A Go library for securely retrieving secrets from multiple secret vaults or backends.

Supported backends in this alpha version:
- AWS Secrets Manager

## Features
- Retrieve secrets from AWS Secrets Manager and other backends
- Extensible for defining other vault/backend services
- Simple API for secret retrieval

## Installation

Add the module to your project:

```
go get github.com/bancolombia/secretsmanager
```

Or, if using Go modules, add to your `go.mod`:

```
require github.com/bancolombia/secretsmanager latest
```

## Configuration

You can configure the library using environment variables, web identity files, or by passing a `Settings` struct.

### Environment Variables

Set AWS credentials and region:

```
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=your-access-key
export AWS_SECRET_ACCESS_KEY=your-secret-key
```

For web identity:
```
export AWS_ROLE_ARN=your-role-arn
export AWS_WEB_IDENTITY_TOKEN_FILE=/path/to/token
```

## Usage

### Initialize the Manager

```go
import (
    "github.com/bancolombia/secretsmanager/api"
    "github.com/bancolombia/secretsmanager"
)

awsopts := make(map[string]interface{})
awsopts["region"] = "us-east-1"
settings := api.Settings{
    VaultType: VaultTypeAwsSecretManager, // AWS Secrets Manager
    VaultConfig: awsopts,
}
manager := secretsmanager.NewSecretsManager(settings)
```

### Retrieve a Secret

```go
secret, err := manager.PullSecret("my-secret-key")
if err != nil {
    // handle error
}
fmt.Println("Secret:", secret)
```

## Testing

Run unit tests:

```
go test ./...
```

## Contributing

Contributions are welcome! Please open issues or submit pull requests.

## License

MIT License
