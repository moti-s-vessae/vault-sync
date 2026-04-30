package vault

import "errors"

// ErrAccessDenied is returned when a Vault policy check fails for a given path.
var ErrAccessDenied = errors.New("vault: access denied by policy")

// ErrSecretNotFound is returned when a secret path does not exist in Vault.
var ErrSecretNotFound = errors.New("vault: secret not found")

// ErrVaultSealed is returned when the Vault instance is sealed.
var ErrVaultSealed = errors.New("vault: instance is sealed")

// ErrVaultUninitialized is returned when the Vault instance is not yet initialized.
var ErrVaultUninitialized = errors.New("vault: instance is uninitialized")

// ErrCacheExpired is returned when a cached entry has passed its TTL.
var ErrCacheExpired = errors.New("vault: cache entry expired")

// ErrInvalidAddress is returned when the Vault address is malformed.
var ErrInvalidAddress = errors.New("vault: invalid server address")
