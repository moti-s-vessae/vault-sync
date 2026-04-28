# vault-sync

> CLI tool to sync secrets from HashiCorp Vault to local `.env` files with namespace filtering

---

## Installation

```bash
go install github.com/yourusername/vault-sync@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/vault-sync.git
cd vault-sync
go build -o vault-sync .
```

---

## Usage

Set your Vault address and token, then run:

```bash
export VAULT_ADDR="https://vault.example.com"
export VAULT_TOKEN="s.your-token-here"

# Sync all secrets under a namespace to a .env file
vault-sync --namespace secret/myapp --output .env

# Filter by environment
vault-sync --namespace secret/myapp/production --output .env.production
```

**Example output (`.env`):**
```
DB_HOST=db.example.com
DB_PASSWORD=supersecret
API_KEY=abc123
```

### Flags

| Flag | Description | Default |
|-------------|-------------------------------|---------|
| `--namespace` | Vault secret path/namespace | required |
| `--output` | Output `.env` file path | `.env` |
| `--overwrite` | Overwrite existing file | `false` |
| `--addr` | Vault server address | `$VAULT_ADDR` |

---

## Requirements

- Go 1.21+
- HashiCorp Vault with KV v2 secrets engine

---

## License

[MIT](LICENSE) © 2024 yourusername