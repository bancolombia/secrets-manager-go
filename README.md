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
go get github.com/bancolombia/secrets-manager-go
```

Or, if using Go modules, add to your `go.mod`:

```
require github.com/bancolombia/secrets-manager-go latest
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

### Initialize the manager

#### For secrets stored in AWS SM

```go
import (
    "github.com/bancolombia/secrets-manager-go/api"
    "github.com/bancolombia/secrets-manager-go"
)

// initialize the manager
awsopts := make(map[string]interface{})
awsopts["region"] = "us-east-1"
settings := api.Settings{
    VaultType: secretsmanager.VaultTypeAwsSecretManager, // AWS Secrets Manager
    VaultConfig: awsopts,
}
manager := secretsmanager.NewSecretsManager(settings)
```

#### Secrets injected as Environment Variables

For components where secrets are injected as environment variables (e.g. by the External Secrets Operator or Kubernetes `envFrom`):

```go
settings := api.Settings{
    VaultType: secretsmanager.VaultTypeEnv,
}
manager := secretsmanager.NewSecretsManager(settings)
```

See [Secret name resolution](#secret-name-resolution) for how `PullSecret` maps the requested key to an environment variable name.

#### Secrets Mounted as Files

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

See [Secret name resolution](#secret-name-resolution) for how `PullSecret` maps the requested key to a file on disk.

### Secret name resolution

The name you pass to `PullSecret(name)` is mapped to a different physical location depending on the backend. Each backend applies its own rules, summarized below.

#### AWS Secrets Manager (`VaultTypeAwsSecretManager`)

The name is forwarded to AWS verbatim as the `SecretId` of a `GetSecretValue` call — no normalization is applied. You can pass either the secret's friendly name (e.g. `prod/db/password`) or its full ARN; both are accepted by the AWS SDK.

| `PullSecret(...)` | AWS `SecretId` |
| ----------------- | -------------- |
| `"prod/db/password"` | `prod/db/password` |
| `"arn:aws:secretsmanager:us-east-1:123456789012:secret:prod/db/password-AbCdEf"` | same ARN, verbatim |

#### Environment variables (`VaultTypeEnv`)

Resolution is a two-step fallback:

1. The exact name is looked up with `os.LookupEnv`. If a variable with that name exists, its value is returned as-is.
2. Otherwise, the name is **normalized** by (a) uppercasing it and (b) replacing every `-` and `.` with `_`. The normalized form is then looked up. If it also misses, an error is returned naming both forms that were tried.

Only `-` and `.` are rewritten — any other character (including `/`, `:`, digits, underscores) is left alone and is matched verbatim against the environment. This matches how Kubernetes `envFrom` and the External Secrets Operator derive env var names from secret keys.

| `PullSecret(...)` | Env vars tried, in order |
| ----------------- | ------------------------ |
| `"my-secret"` | `my-secret`, then `MY_SECRET` |
| `"app.db.password"` | `app.db.password`, then `APP_DB_PASSWORD` |
| `"MY_SECRET"` | `MY_SECRET` only (already normalized — no second lookup) |
| `"db/pass"` | `db/pass`, then `DB/PASS` (both will typically miss — Kubernetes exposes this key as `DB_PASS`, so pull it as `DB_PASS` or `db-pass`) |

#### Mounted files (`VaultTypeFile`)

The name is used verbatim as the filename under the configured mount path (default `/mnt/secrets-store`). The library:

- Rejects empty names, names containing path separators (`/`, `\`), and names containing `..`, to prevent path-traversal reads outside the mount point.
- Reads `<path>/<name>` with `os.ReadFile`.
- Returns the file's contents with trailing `\r` and `\n` characters trimmed (spaces and tabs are preserved).

| `VaultConfig["path"]` | `PullSecret(...)` | File read |
| --------------------- | ----------------- | --------- |
| `/mnt/secrets-store` (default) | `"my-secret"` | `/mnt/secrets-store/my-secret` |
| `/etc/app/secrets` | `"db.password"` | `/etc/app/secrets/db.password` |
| any | `"../etc/passwd"` | — rejected with an invalid-name error |
| any | `"sub/key"` | — rejected with an invalid-name error |

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
