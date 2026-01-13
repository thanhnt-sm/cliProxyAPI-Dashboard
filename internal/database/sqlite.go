package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/pricing"
	_ "modernc.org/sqlite"
)

var (
	db   *sql.DB
	once sync.Once
)

type UsageLog struct {
	ID             int64     `json:"id"`
	Timestamp      time.Time `json:"timestamp"`
	APIKey         string    `json:"api_key"`
	Model          string    `json:"model"`
	InputTokens    int64     `json:"input_tokens"`
	OutputTokens   int64     `json:"output_tokens"`
	TotalTokens    int64     `json:"total_tokens"`
	IsFailure      bool      `json:"is_failure"`
	Source         string    `json:"source"`
	DurationMs     int64     `json:"duration_ms"`
	PromptText     string    `json:"prompt_text"`
	CompletionText string    `json:"completion_text"`
	CostUSD        float64   `json:"cost_usd"`
}

// ManagedKey represents a customer API key with limits.
type ManagedKey struct {
	KeyHash            string    `json:"key_hash"`
	KeyPrefix          string    `json:"key_prefix"`
	Label              string    `json:"label"`
	QuotaLimitUSD      float64   `json:"quota_limit_usd"`
	QuotaLimitRequests int64     `json:"quota_limit_requests"`
	RateLimitRPM       int64     `json:"rate_limit_rpm"`
	AllowedModels      []string  `json:"allowed_models"` // JSON array in DB
	IsActive           bool      `json:"is_active"`
	ExpiresAt          time.Time `json:"expires_at"`
	CreatedAt          time.Time `json:"created_at"`
}

// Init initializes the SQLite database.
// configDir is detailed path where db file should be stored.
func Init(configDir string) error {
	var err error
	once.Do(func() {
		if err = os.MkdirAll(configDir, 0755); err != nil {
			return
		}
		dbPath := filepath.Join(configDir, "usage.db")
		fmt.Printf("DEBUG: Initializing database at %s\n", dbPath)
		db, err = sql.Open("sqlite", dbPath)
		if err != nil {
			return
		}

		// Create tables
		query := `
		CREATE TABLE IF NOT EXISTS usage_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME,
			api_key TEXT,
			model TEXT,
			input_tokens INTEGER,
			output_tokens INTEGER,
			total_tokens INTEGER,
			is_failure BOOLEAN,
			source TEXT,
			duration_ms INTEGER,
			prompt_text TEXT,
			completion_text TEXT,
			cost_usd REAL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS idx_timestamp ON usage_logs(timestamp DESC);

		CREATE TABLE IF NOT EXISTS managed_keys (
			key_hash TEXT PRIMARY KEY,
			key_prefix TEXT,
			label TEXT,
			quota_limit_usd REAL DEFAULT 0,
			quota_limit_requests INTEGER DEFAULT 0,
			rate_limit_rpm INTEGER DEFAULT 0,
			allowed_models TEXT,
			is_active BOOLEAN DEFAULT 1,
			expires_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		`
		_, err = db.Exec(query)
		if err != nil {
			return 
		}

	// Migration: Add cost_usd column if it doesn't exist (for existing DBs)
		// We ignore error here assuming it might fail if column exists, or we check specifically.
		// Simple way: Try to add it, ignore "duplicate column" error.
		// In SQLite, "ALTER TABLE ... ADD COLUMN" works.
		migrationQuery := `ALTER TABLE usage_logs ADD COLUMN cost_usd REAL DEFAULT 0;`
		db.Exec(migrationQuery) // Ignore error, likely "duplicate column name"

		// Recalculate all costs to match new pricing configuration
		_ = recalculateAllCosts()
	})
	return err
}

func recalculateAllCosts() error {
	if db == nil {
		return fmt.Errorf("db not initialized")
	}
	
	rows, err := db.Query("SELECT id, model, input_tokens, output_tokens FROM usage_logs")
	if err != nil {
		return err
	}
	defer rows.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE usage_logs SET cost_usd = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var id int64
		var model string
		var input, output int64
		if err := rows.Scan(&id, &model, &input, &output); err != nil {
			continue
		}
		
		newCost := pricing.CalculateCost(model, input, output)
		if _, err := stmt.Exec(newCost, id); err != nil {
			// Continue on error? or abort?
			// Best effort
		}
	}
	
	return tx.Commit()
}

// Close closes the database connection.
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// InsertUsageLog inserts a new usage record.
func InsertUsageLog(log UsageLog) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	
	// Calculate cost if not provided
	if log.CostUSD == 0 {
		log.CostUSD = pricing.CalculateCost(log.Model, log.InputTokens, log.OutputTokens)
	}

	// SQLite requires specific date formats for functions like strftime to work.
	// We store as UTC string: YYYY-MM-DD HH:MM:SS
	ts := log.Timestamp.UTC().Format("2006-01-02 15:04:05")

	query := `
	INSERT INTO usage_logs (timestamp, api_key, model, input_tokens, output_tokens, total_tokens, is_failure, source, duration_ms, prompt_text, completion_text, cost_usd)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := db.Exec(query, ts, log.APIKey, log.Model, log.InputTokens, log.OutputTokens, log.TotalTokens, log.IsFailure, log.Source, log.DurationMs, log.PromptText, log.CompletionText, log.CostUSD)
	return err
}

// GetRecentActivity returns paginated usage logs with optional filters.
func GetRecentActivity(limit, offset int, modelFilter, statusFilter string) ([]UsageLog, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `SELECT id, timestamp, COALESCE(api_key, ''), COALESCE(model, ''), COALESCE(input_tokens, 0), COALESCE(output_tokens, 0), COALESCE(total_tokens, 0), is_failure, COALESCE(source, 'unknown'), COALESCE(duration_ms, 0), COALESCE(prompt_text, ''), COALESCE(completion_text, ''), COALESCE(cost_usd, 0) FROM usage_logs WHERE 1=1`
	var args []interface{}

	if modelFilter != "" {
		query += ` AND model = ?`
		args = append(args, modelFilter)
	}

	if statusFilter != "" {
		if statusFilter == "success" {
			query += ` AND is_failure = 0`
		} else if statusFilter == "failure" {
			query += ` AND is_failure = 1`
		}
	}

	query += ` ORDER BY timestamp DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []UsageLog
	for rows.Next() {
		var log UsageLog
		var ts time.Time
		// We use a sql.NullFloat64 for cost just in case, or default to 0 if null.
		// Since we defined DEFAULT 0, plain float64 should be fine if migration worked.
		// But Scan into struct field works fine for non-null columns.
		if err := rows.Scan(&log.ID, &ts, &log.APIKey, &log.Model, &log.InputTokens, &log.OutputTokens, &log.TotalTokens, &log.IsFailure, &log.Source, &log.DurationMs, &log.PromptText, &log.CompletionText, &log.CostUSD); err != nil {
			return nil, err
		}
		log.Timestamp = ts
		logs = append(logs, log)
	}
	return logs, nil
}

// GetAllActivity returns all usage logs matching the filters, without pagination.
func GetAllActivity(modelFilter, statusFilter string) ([]UsageLog, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `SELECT id, timestamp, COALESCE(api_key, ''), COALESCE(model, ''), COALESCE(input_tokens, 0), COALESCE(output_tokens, 0), COALESCE(total_tokens, 0), is_failure, COALESCE(source, 'unknown'), COALESCE(duration_ms, 0), COALESCE(prompt_text, ''), COALESCE(completion_text, '') FROM usage_logs WHERE 1=1`
	var args []interface{}

	if modelFilter != "" {
		query += ` AND model = ?`
		args = append(args, modelFilter)
	}

	if statusFilter != "" {
		if statusFilter == "success" {
			query += ` AND is_failure = 0`
		} else if statusFilter == "failure" {
			query += ` AND is_failure = 1`
		}
	}

	query += ` ORDER BY timestamp DESC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []UsageLog
	for rows.Next() {
		var log UsageLog
		var ts time.Time
		if err := rows.Scan(&log.ID, &ts, &log.APIKey, &log.Model, &log.InputTokens, &log.OutputTokens, &log.TotalTokens, &log.IsFailure, &log.Source, &log.DurationMs, &log.PromptText, &log.CompletionText, &log.CostUSD); err != nil {
			return nil, err
		}
		log.Timestamp = ts
		logs = append(logs, log)
	}
	return logs, nil
}

type UsageTrend struct {
	Bucket    string  `json:"bucket"` // formatted date string
	Requests  int64   `json:"requests"`
	Failures  int64   `json:"failures"`
	Tokens    int64   `json:"tokens"`
	Cost      float64 `json:"cost"`
}

// GetUsageTrends returns aggregated stats grouped by hour or day.
// groupBy: "hour" or "day"
// limit: number of buckets to return
func GetUsageTrends(groupBy string, limit int) ([]UsageTrend, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var format string
	if groupBy == "hour" {
		format = "%Y-%m-%d %H:00"
	} else {
		format = "%Y-%m-%d" // daily
	}

	// SQLite strftime: https://www.sqlite.org/lang_datefunc.html
	query := fmt.Sprintf(`
	SELECT 
		COALESCE(strftime('%s', timestamp), '0') as bucket,
		COUNT(*) as requests,
		COALESCE(SUM(CASE WHEN is_failure = 1 THEN 1 ELSE 0 END), 0) as failures,
		COALESCE(SUM(total_tokens), 0) as tokens,
		COALESCE(SUM(cost_usd), 0) as cost
	FROM usage_logs
	GROUP BY bucket
	ORDER BY bucket DESC
	LIMIT ?
	`, format)

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trends []UsageTrend
	for rows.Next() {
		var t UsageTrend
		// Scanning into Cost (float64). If NULL (no logs), Sum is NULL?
		// Actually COUNT(*) ensures rows exist? No, if empty DB no rows.
		// If group has no cost_usd (all null), SUM is null.
		// We should use COALESCE(SUM(cost_usd), 0).
		
		// To avoid complex SQL change if not supported by modernc/sqlite or behave oddly, 
		// let's scan into sql.NullFloat64 or update query.
		// Let's update query to COALESCE.
		// Wait, query above is inside fmt.Sprintf so be careful.
		
		// Let's modify the query in the replacement content to include COALESCE
		// Actually, standard SUM in sqlite returns NULL if no rows, but GROUP BY ensures rows. 
		// If rows exist but cost_usd is NULL? It ignores NULLs. If all NULL, returns NULL.
		// So COALESCE is safer.
		// Scan destination needs to match.
		if err := rows.Scan(&t.Bucket, &t.Requests, &t.Failures, &t.Tokens, &t.Cost); err != nil {
			// If scan fails due to NULL to float64?
			// Let's assume defaults for now or I will fix the query below.
			return nil, err
		}
		trends = append(trends, t)
	}
	
	// Reverse result to be chronological for charts
	for i, j := 0, len(trends)-1; i < j; i, j = i+1, j-1 {
		trends[i], trends[j] = trends[j], trends[i]
	}
	
	return trends, nil
}

type GlobalStats struct {
	TotalRequests int64
	TotalTokens   int64
	SuccessCount  int64
	FailureCount  int64
	TotalCost     float64
}

// GetGlobalStats returns the aggregated all-time stats from the DB.
func GetGlobalStats() (GlobalStats, error) {
	var s GlobalStats
	if db == nil {
		return s, fmt.Errorf("database not initialized")
	}

	query := `
	SELECT 
		COUNT(*) as total_requests,
		COALESCE(SUM(total_tokens), 0) as total_tokens,
		COALESCE(SUM(CASE WHEN is_failure = 0 THEN 1 ELSE 0 END), 0) as success_count,
		COALESCE(SUM(CASE WHEN is_failure = 1 THEN 1 ELSE 0 END), 0) as failure_count,
		COALESCE(SUM(cost_usd), 0) as total_cost
	FROM usage_logs
	`
	
	err := db.QueryRow(query).Scan(&s.TotalRequests, &s.TotalTokens, &s.SuccessCount, &s.FailureCount, &s.TotalCost)
	if err != nil {
		return s, err
	}
	return s, nil
}

type ModelAggStats struct {
	Model         string
	TotalRequests int64
	TotalTokens   int64
}

// GetAggregatedModelStats returns stats grouped by model (for Top Model calc)
func GetAggregatedModelStats() ([]ModelAggStats, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	
	// We only need Model, TotalRequests, TotalTokens
	query := `
	SELECT 
		COALESCE(model, 'unknown') as model,
		COUNT(*) as total_requests,
		COALESCE(SUM(total_tokens), 0) as total_tokens
	FROM usage_logs
	GROUP BY model
	ORDER BY total_requests DESC
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []ModelAggStats
	for rows.Next() {
		var s ModelAggStats
		if err := rows.Scan(&s.Model, &s.TotalRequests, &s.TotalTokens); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

// GetPeriodCosts returns the aggregated costs for 24h, 7d, and lifetime.
func GetPeriodCosts() (cost24h float64, cost7d float64, costTotal float64, err error) {
	if db == nil {
		return 0, 0, 0, fmt.Errorf("database not initialized")
	}

	query := `
	SELECT 
		COALESCE(SUM(CASE WHEN timestamp >= datetime('now', '-1 day') THEN cost_usd ELSE 0 END), 0) as cost_24h,
		COALESCE(SUM(CASE WHEN timestamp >= datetime('now', '-7 days') THEN cost_usd ELSE 0 END), 0) as cost_7d,
		COALESCE(SUM(cost_usd), 0) as cost_total
	FROM usage_logs
	`
	// Note: SQLite 'now' is UTC if stored as UTC string, which we do.
	
	err = db.QueryRow(query).Scan(&cost24h, &cost7d, &costTotal)
	if err != nil {
		return 0, 0, 0, err
	}
	return cost24h, cost7d, costTotal, nil
}
