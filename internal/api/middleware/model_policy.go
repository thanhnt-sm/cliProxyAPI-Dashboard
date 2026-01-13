package middleware

import (
	"bytes"

	"io"
	"net/http"


	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/database"
	"github.com/tidwall/gjson"
)

// EnforceModelPolicy checks if the requested model is allowed for the authenticated managed key.
// It inspects the request body for "model" field.
func EnforceModelPolicy() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Check if managed key exists in context (set by AuthMiddleware)
		val, exists := c.Get("managedKey")
		if !exists {
			c.Next()
			return
		}

		managedKey, ok := val.(*database.ManagedKey)
		if !ok || managedKey == nil {
			c.Next()
			return
		}

		// If no allowed models are defined (empty list), allow all
		if len(managedKey.AllowedModels) == 0 {
			c.Next()
			return
		}

		// 2. Read body to extract model
		// We use ShouldBindBodyWith or manually read/restore to ensure body triggers EOF for next handler
		// But since the next handler (OpenAI generic) might use GetRawData, we must be careful.
		// Safe approach: Read, Restore.

		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
			return
		}

		// Restore body immediately
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// 3. Extract model
		// Handle both JSON body (standard) and potentially other formats if needed, but OpenAI API dictates JSON.
		// Use gjson for efficient extraction without full unmarshal
		model := gjson.GetBytes(bodyBytes, "model").String()

		if model == "" {
			// If model is missing in body, it might be a problem for the handler too,
			// or it's a request that defaults.
			// Let's assume validation passes if model is not specified (e.g. list models)?
			// No, this middleware should only be applied to endpoints that REQUIRE model (Completions/Chat).
			// If applied to ListModels, body is empty.
			c.Next()
			return
		}

		// 4. Validate against allowed models
		allowed := false
		for _, m := range managedKey.AllowedModels {
			if m == model {
				allowed = true
				break
			}
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"message": "Model '" + model + "' is not allowed for this API key",
					"type":    "permission_error",
					"code":    "model_not_allowed",
				},
			})
			return
		}

		c.Next()
	}
}

// FilterModelsList filters the response of /v1/models to only show allowed models.
// This requires capturing the response body, which is expensive and complex with Gin.
// A better approach is to modify the handler, but we want to avoid SDK coupling.
// Alternative: Put this logic in the handler wrapper in server.go, passed as a config or callback.
// For now, we will skip filtering /v1/models response and focus on ENFORCEMENT.
