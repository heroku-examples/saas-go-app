package api

import (
	"net/http"

	"saas-go-app/internal/db"

	"github.com/gin-gonic/gin"
)

// AnalyticsResponse represents analytics data
type AnalyticsResponse struct {
	TotalCustomers   int     `json:"total_customers"`
	TotalAccounts    int     `json:"total_accounts"`
	ActiveAccounts   int     `json:"active_accounts"`
	InactiveAccounts int     `json:"inactive_accounts"`
	AvgAccountsPerCustomer float64 `json:"avg_accounts_per_customer"`
}

// GetAnalytics retrieves analytics data from the follower pool
// @Summary      Get analytics overview
// @Description  Get overall analytics statistics including customer and account counts
// @Tags         analytics
// @Accept       json
// @Produce      json
// @Success      200  {object}  AnalyticsResponse
// @Failure      500  {object}  map[string]string
// @Router       /analytics [get]
// @Security     BearerAuth
func GetAnalytics(c *gin.Context) {
	// Use analytics DB (follower pool) for read-only analytics queries
	analyticsDB := db.AnalyticsDB
	if analyticsDB == nil {
		analyticsDB = db.PrimaryDB
	}

	var totalCustomers int
	err := analyticsDB.QueryRow("SELECT COUNT(*) FROM customers").Scan(&totalCustomers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch customer count"})
		return
	}

	var totalAccounts int
	err = analyticsDB.QueryRow("SELECT COUNT(*) FROM accounts").Scan(&totalAccounts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch account count"})
		return
	}

	var activeAccounts int
	err = analyticsDB.QueryRow("SELECT COUNT(*) FROM accounts WHERE status = 'active'").Scan(&activeAccounts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch active account count"})
		return
	}

	var inactiveAccounts int
	err = analyticsDB.QueryRow("SELECT COUNT(*) FROM accounts WHERE status = 'inactive'").Scan(&inactiveAccounts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch inactive account count"})
		return
	}

	var avgAccountsPerCustomer float64
	if totalCustomers > 0 {
		err = analyticsDB.QueryRow(
			"SELECT COALESCE(AVG(account_count), 0) FROM (SELECT customer_id, COUNT(*) as account_count FROM accounts GROUP BY customer_id) AS subquery",
		).Scan(&avgAccountsPerCustomer)
		if err != nil {
			avgAccountsPerCustomer = 0
		}
	}

	response := AnalyticsResponse{
		TotalCustomers:        totalCustomers,
		TotalAccounts:         totalAccounts,
		ActiveAccounts:        activeAccounts,
		InactiveAccounts:      inactiveAccounts,
		AvgAccountsPerCustomer: avgAccountsPerCustomer,
	}

	c.JSON(http.StatusOK, response)
}

// GetCustomerAnalytics retrieves analytics for a specific customer
// @Summary      Get customer analytics
// @Description  Get analytics for a specific customer including account counts
// @Tags         analytics
// @Accept       json
// @Produce      json
// @Param        customer_id  path      string  true  "Customer ID"
// @Success      200          {object}  map[string]interface{}
// @Failure      500          {object}  map[string]string
// @Router       /analytics/customers/{customer_id} [get]
// @Security     BearerAuth
func GetCustomerAnalytics(c *gin.Context) {
	customerID := c.Param("customer_id")

	analyticsDB := db.AnalyticsDB
	if analyticsDB == nil {
		analyticsDB = db.PrimaryDB
	}

	var accountCount int
	var activeCount int
	err := analyticsDB.QueryRow(
		"SELECT COUNT(*), COUNT(CASE WHEN status = 'active' THEN 1 END) FROM accounts WHERE customer_id = $1",
		customerID,
	).Scan(&accountCount, &activeCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch customer analytics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"customer_id":   customerID,
		"total_accounts": accountCount,
		"active_accounts": activeCount,
		"inactive_accounts": accountCount - activeCount,
	})
}

