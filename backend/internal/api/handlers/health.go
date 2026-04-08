package handlers

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	GoVersion string `json:"go_version"`
	Goroutine int    `json:"goroutine"`
}

// HealthCheck returns server health status
// @Summary Health check
// @Description Check if the server is running
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:    "ok",
		Version:   "0.1.0",
		GoVersion: runtime.Version(),
		Goroutine: runtime.NumGoroutine(),
	})
}