package handlers

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
)

type HealthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	GoVersion string `json:"go_version"`
	Goroutine int    `json:"goroutine"`
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:    "ok",
		Version:   "0.1.0",
		GoVersion: runtime.Version(),
		Goroutine: runtime.NumGoroutine(),
	})
}