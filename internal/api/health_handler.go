package api

import (
	"net/http"

	"saas-go-app/internal/db"

	"github.com/gin-gonic/gin"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status      string `json:"status"`
	Database    string `json:"database"`
	AnalyticsDB string `json:"analytics_db"`
}

// HealthCheck performs a health check on the service
// @Summary      Health check
// @Description  Check the health status of the service and database connections
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200  {object}  HealthResponse
// @Failure      503  {object}  HealthResponse
// @Router       /health [get]
func HealthCheck(c *gin.Context) {
	response := HealthResponse{
		Status: "healthy",
	}

	// Check primary database
	if err := db.PrimaryDB.Ping(); err != nil {
		response.Status = "unhealthy"
		response.Database = "disconnected"
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}
	response.Database = "connected"

	// Check analytics database
	if db.AnalyticsDB != nil && db.AnalyticsDB != db.PrimaryDB {
		if err := db.AnalyticsDB.Ping(); err != nil {
			response.AnalyticsDB = "disconnected"
		} else {
			response.AnalyticsDB = "connected"
		}
	} else {
		response.AnalyticsDB = "using primary"
	}

	if response.Status == "healthy" {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusServiceUnavailable, response)
	}
}

