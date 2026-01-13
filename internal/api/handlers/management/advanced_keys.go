package management

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/database"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/registry"
)

// ListRegistryModels returns all available models from the global registry.
func (h *Handler) ListRegistryModels(c *gin.Context) {
	reg := registry.GetGlobalRegistry()
	models := reg.GetAvailableModels("openai")
	c.JSON(http.StatusOK, gin.H{"models": models})
}

// ListManagedKeys returns all managed keys.
func (h *Handler) ListManagedKeys(c *gin.Context) {
	keys, err := database.ListManagedKeys()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list keys: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"keys": keys})
}

// CreateManagedKeyRequest defines the body for creating a key.
type CreateManagedKeyRequest struct {
	Label              string   `json:"label" binding:"required"`
	QuotaLimitUSD      float64  `json:"quota_limit_usd"`      // Optional, 0 = unlimited
	QuotaLimitRequests int64    `json:"quota_limit_requests"` // Optional
	RateLimitRPM       int64    `json:"rate_limit_rpm"`       // Optional
	AllowedModels      []string `json:"allowed_models"`       // Optional, empty = all
	ExpiresInSeconds   int64    `json:"expires_in_seconds"`   // Optional
}

// CreateManagedKey creates a new API key.
func (h *Handler) CreateManagedKey(c *gin.Context) {
	var req CreateManagedKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Generate a secure random key
	// Format: sk-managed-<random_hex> (total 40 chars)
	rawKeyBytes := make([]byte, 24)
	if _, err := rand.Read(rawKeyBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate key"})
		return
	}
	rawKey := "sk-managed-" + hex.EncodeToString(rawKeyBytes)

	// Hash it for storage
	hash := sha256.Sum256([]byte(rawKey))
	hashStr := hex.EncodeToString(hash[:])

	// Calculate Expiration
	var expiresAt time.Time
	if req.ExpiresInSeconds > 0 {
		expiresAt = time.Now().Add(time.Duration(req.ExpiresInSeconds) * time.Second)
	}

	// Create Struct
	k := database.ManagedKey{
		KeyHash:            hashStr,
		KeyPrefix:          rawKey[:15] + "...", // Show only prefix
		Label:              req.Label,
		QuotaLimitUSD:      req.QuotaLimitUSD,
		QuotaLimitRequests: req.QuotaLimitRequests,
		RateLimitRPM:       req.RateLimitRPM,
		AllowedModels:      req.AllowedModels,
		IsActive:           true,
		ExpiresAt:          expiresAt,
	}

	if err := database.InsertManagedKey(k); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save key: " + err.Error()})
		return
	}

	// Return the raw key ONLY ONCE
	c.JSON(http.StatusCreated, gin.H{
		"api_key": rawKey,
		"details": k,
		"warning": "This key will not be shown again. Save it now.",
	})
}

// DeleteManagedKey revokes a key.
func (h *Handler) DeleteManagedKey(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing key hash"})
		return
	}
	if err := database.DeleteManagedKey(hash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete key: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
