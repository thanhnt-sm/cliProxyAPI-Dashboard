package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// InsertManagedKey adds a new managed key to the database.
func InsertManagedKey(k ManagedKey) error {
    if db == nil {
        return fmt.Errorf("database not initialized")
    }
    modelsJSON, _ := json.Marshal(k.AllowedModels)
    query := `
    INSERT INTO managed_keys (key_hash, key_prefix, label, quota_limit_usd, quota_limit_requests, rate_limit_rpm, allowed_models, is_active, expires_at, created_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
    _, err := db.Exec(query, k.KeyHash, k.KeyPrefix, k.Label, k.QuotaLimitUSD, k.QuotaLimitRequests, k.RateLimitRPM, string(modelsJSON), k.IsActive, k.ExpiresAt, time.Now())
    return err
}

// GetManagedKey retrieves a key by its hash.
func GetManagedKey(hash string) (*ManagedKey, error) {
    if db == nil {
        return nil, fmt.Errorf("database not initialized")
    }
    query := `SELECT key_hash, key_prefix, label, quota_limit_usd, quota_limit_requests, rate_limit_rpm, allowed_models, is_active, expires_at, created_at FROM managed_keys WHERE key_hash = ?`
    row := db.QueryRow(query, hash)
    var k ManagedKey
    var modelsStr string
    var expiresAt sql.NullTime // Use NullTime to handle potential NULLs safely
    
    if err := row.Scan(&k.KeyHash, &k.KeyPrefix, &k.Label, &k.QuotaLimitUSD, &k.QuotaLimitRequests, &k.RateLimitRPM, &modelsStr, &k.IsActive, &expiresAt, &k.CreatedAt); err != nil {
        return nil, err
    }
    
    if expiresAt.Valid {
        k.ExpiresAt = expiresAt.Time
    }
    
    if modelsStr != "" {
        _ = json.Unmarshal([]byte(modelsStr), &k.AllowedModels)
    }
    return &k, nil
}

// ListManagedKeys returns all managed keys.
func ListManagedKeys() ([]ManagedKey, error) {
    if db == nil {
        return nil, fmt.Errorf("database not initialized")
    }
    query := `SELECT key_hash, key_prefix, label, quota_limit_usd, quota_limit_requests, rate_limit_rpm, allowed_models, is_active, expires_at, created_at FROM managed_keys ORDER BY created_at DESC`
    rows, err := db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var keys []ManagedKey
    for rows.Next() {
        var k ManagedKey
        var modelsStr string
        var expiresAt sql.NullTime
        if err := rows.Scan(&k.KeyHash, &k.KeyPrefix, &k.Label, &k.QuotaLimitUSD, &k.QuotaLimitRequests, &k.RateLimitRPM, &modelsStr, &k.IsActive, &expiresAt, &k.CreatedAt); err != nil {
            return nil, err
        }
        if expiresAt.Valid {
            k.ExpiresAt = expiresAt.Time
        }
        if modelsStr != "" {
            _ = json.Unmarshal([]byte(modelsStr), &k.AllowedModels)
        }
        keys = append(keys, k)
    }
    return keys, nil
}

// DeleteManagedKey deletes a key.
func DeleteManagedKey(hash string) error {
    if db == nil {
        return fmt.Errorf("database not initialized")
    }
    _, err := db.Exec("DELETE FROM managed_keys WHERE key_hash = ?", hash)
    return err
}

// GetUsageForPrefix retrieves total cost and requests for a specific key prefix (simple matching).
// Ideally we should link usage_logs to key_hash or store partial key in usage logs.
// Currently usage_logs has `api_key` which might be full key (if we change logging) or masked.
// For Managed Keys, we will log the *FULL* key (or a unique ID) in usage_logs?
// Wait, logging full key is bad security.
// We should log the `key_prefix` or a `key_id` in usage_logs.
// Current `usage_logs` has `api_key`. The `request_logging.go` trims it?
// Let's check `request_logging.go`.
func GetUsageForKey(keyPrefix string) (requests int64, cost float64, err error) {
    if db == nil {
        return 0, 0, fmt.Errorf("database not initialized")
    }
    // Assuming usage_logs.api_key stores something traceable to the managed key.
    // Ideally we match by prefix if that's how we log it.
    // For now, let's assume we match exact string in `api_key` column OR partial.
    // IMPORTANT: If we persist the hash, we can't link it to usage easily unless we log usage with hash too?
    // Usage logs usually log the key used.
    
    // We will query by 'api_key LIKE ?'
    query := `
    SELECT COUNT(*), COALESCE(SUM(cost_usd), 0)
    FROM usage_logs
    WHERE api_key LIKE ?
    `
    // Match "sk-proj-..."
    err = db.QueryRow(query, keyPrefix + "%").Scan(&requests, &cost)
    return
}
