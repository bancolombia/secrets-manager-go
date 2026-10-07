# SecretsManager Go - Bancolombia (alpha)

A Go library for securely retrieving secrets from multiple secret vaults or backends.

Supported backends in this alpha version:
- AWS Secrets Manager
- Environment variables (secrets synced by the External Secrets Operator or injected with Kubernetes `envFrom`)
- Mounted files (secrets exposed by the Secrets Store CSI Driver or as secret volumes)

## Features
- Retrieve secrets from AWS Secrets Manager, environment variables, or mounted files
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
    VaultType: secretsmanager.VaultTypeAwsSecretManager, // AWS Secrets Manager
    VaultConfig: awsopts,
}
manager := secretsmanager.NewSecretsManager(settings)
```

### Environment Variables

For components where secrets are injected as environment variables (e.g. by the External Secrets Operator or Kubernetes `envFrom`):

```go
settings := api.Settings{
    VaultType: secretsmanager.VaultTypeEnv,
}
manager := secretsmanager.NewSecretsManager(settings)
```

Lookup rules: the exact name is checked first; if not present, the name is uppercased with `-` and `.` replaced by `_`, so `PullSecret("my-secret")` finds `MY_SECRET`. Only those two characters are normalized: a Kubernetes secret key like `db/pass` is exposed by `envFrom` as `DB_PASS`, so it must be pulled as `DB_PASS`.

### Mounted Files

For components where secrets are exposed as files (e.g. by the Secrets Store CSI Driver or a secret volume):

```go
fileopts := make(map[string]interface{})
fileopts["path"] = "/mnt/secrets-store" // optional, this is the default
settings := api.Settings{
    VaultType: secretsmanager.VaultTypeFile,
    VaultConfig: fileopts,
}
manager := secretsmanager.NewSecretsManager(settings)
```

Each secret maps to a file named after the secret key under the configured mount path, so `PullSecret("my-secret")` reads `<path>/my-secret`. Trailing newlines are trimmed from the file contents, and secret names must not contain path separators or `..`.

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

Or run the full CI-equivalent checks (format, vet, build, race tests with coverage):

```
make ci
```

## Contributing

Contributions are welcome! Please open issues or submit pull requests.

## License

MIT License
