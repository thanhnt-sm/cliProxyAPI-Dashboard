package management

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/database"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/usage"
)

type usageExportPayload struct {
	Version    int                      `json:"version"`
	ExportedAt time.Time                `json:"exported_at"`
	Usage      usage.StatisticsSnapshot `json:"usage"`
}

type usageImportPayload struct {
	Version int                      `json:"version"`
	Usage   usage.StatisticsSnapshot `json:"usage"`
}

// GetUsageStatistics returns the in-memory request statistics snapshot.
// GetUsageStatistics returns the in-memory request statistics snapshot.
func (h *Handler) GetUsageStatistics(c *gin.Context) {
	var snapshot usage.StatisticsSnapshot
	if h != nil && h.usageStats != nil {
		snapshot = h.usageStats.Snapshot()
	}

	// Overwrite totals with persistent DB stats to ensure Overview is correct after restart
	if globalStats, err := database.GetGlobalStats(); err == nil {
		snapshot.TotalRequests = globalStats.TotalRequests
		snapshot.TotalTokens = globalStats.TotalTokens
		snapshot.SuccessCount = globalStats.SuccessCount
		snapshot.FailureCount = globalStats.FailureCount
	}

	// Inject Period Costs (24h, 7d, Lifetime)
	if c24, c7d, cTot, err := database.GetPeriodCosts(); err == nil {
		snapshot.Cost24h = c24
		snapshot.Cost7d = c7d
		snapshot.TotalCost = cTot
	}

	// Overwrite per-model stats for "Top Model" logic
	// The frontend looks at snapshot.APIs to find models.
	// We'll synthesize a "GlobalDB" API entry containing all persistent stats.
	if modelStats, err := database.GetAggregatedModelStats(); err == nil && len(modelStats) > 0 {
		// Initialize APIs map if nil
		if snapshot.APIs == nil {
			snapshot.APIs = make(map[string]usage.APISnapshot)
		}
		
		// Create a synthetic API to hold the DB stats so frontend sees them
		// Use a distinct name so it doesn't conflict with in-memory session keys if any
		dbApi := usage.APISnapshot{
			TotalRequests: 0,
			TotalTokens:   0,
			Models:        make(map[string]usage.ModelSnapshot),
		}
		
		for _, ms := range modelStats {
			dbApi.TotalRequests += ms.TotalRequests
			dbApi.TotalTokens += ms.TotalTokens
			
			dbApi.Models[ms.Model] = usage.ModelSnapshot{
				TotalRequests: ms.TotalRequests,
				TotalTokens:   ms.TotalTokens,
				// Details omitted for brevity/performance in overview
			}
		}
		
		// Replace or Merge? 
		// Since we want persistent stats to dominate for "Overview", and in-memory is transient,
		// putting it in a reserved key allows frontend to pick it up if it iterates all.
		// NOTE: Frontend likely iterates ALL APIs. If we add this, it adds to the list.
		// If we want "Top Model" to be correct, we ideally want ONLY this source or merged.
		// Simplest fix: The frontend iterates all APIs and Models. 
		// If we clear existing transient APIs and provide just this Global DB one, it ensures consistency.
		
		snapshot.APIs = map[string]usage.APISnapshot{
			"persistent_db_source": dbApi,
		}
	} else {
		// Log error or fallback
	}

	c.JSON(http.StatusOK, gin.H{
		"usage":           snapshot,
		"failed_requests": snapshot.FailureCount,
	})
}

// ExportUsageStatistics returns a complete usage snapshot for backup/migration.
func (h *Handler) ExportUsageStatistics(c *gin.Context) {
	var snapshot usage.StatisticsSnapshot
	if h != nil && h.usageStats != nil {
		snapshot = h.usageStats.Snapshot()
	}
	c.JSON(http.StatusOK, usageExportPayload{
		Version:    1,
		ExportedAt: time.Now().UTC(),
		Usage:      snapshot,
	})
}

// ImportUsageStatistics merges a previously exported usage snapshot into memory.
func (h *Handler) ImportUsageStatistics(c *gin.Context) {
	if h == nil || h.usageStats == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "usage statistics unavailable"})
		return
	}

	data, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	var payload usageImportPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if payload.Version != 0 && payload.Version != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported version"})
		return
	}

	result := h.usageStats.MergeSnapshot(payload.Usage)
	snapshot := h.usageStats.Snapshot()
	c.JSON(http.StatusOK, gin.H{
		"added":           result.Added,
		"skipped":         result.Skipped,
		"total_requests":  snapshot.TotalRequests,
		"failed_requests": snapshot.FailureCount,
	})
}
