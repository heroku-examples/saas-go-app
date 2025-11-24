package api

import (
	"database/sql"
	"net/http"
	"strconv"

	"saas-go-app/internal/db"
	"saas-go-app/internal/models"

	"github.com/gin-gonic/gin"
)

// GetCustomers retrieves all customers
// @Summary      List all customers
// @Description  Get a list of all customers
// @Tags         customers
// @Accept       json
// @Produce      json
// @Success      200  {array}   models.Customer
// @Failure      500  {object}  map[string]string
// @Router       /customers [get]
// @Security     BearerAuth
func GetCustomers(c *gin.Context) {
	rows, err := db.PrimaryDB.Query(
		"SELECT id, name, email, created_at, updated_at FROM customers ORDER BY created_at DESC",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch customers"})
		return
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		var customer models.Customer
		if err := rows.Scan(&customer.ID, &customer.Name, &customer.Email, &customer.CreatedAt, &customer.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan customer"})
			return
		}
		customers = append(customers, customer)
	}

	c.JSON(http.StatusOK, customers)
}

// GetCustomer retrieves a single customer by ID
// @Summary      Get customer by ID
// @Description  Get a specific customer by their ID
// @Tags         customers
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Customer ID"
// @Success      200  {object}  models.Customer
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /customers/{id} [get]
// @Security     BearerAuth
func GetCustomer(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	var customer models.Customer
	err = db.PrimaryDB.QueryRow(
		"SELECT id, name, email, created_at, updated_at FROM customers WHERE id = $1",
		id,
	).Scan(&customer.ID, &customer.Name, &customer.Email, &customer.CreatedAt, &customer.UpdatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch customer"})
		return
	}

	c.JSON(http.StatusOK, customer)
}

// CreateCustomer creates a new customer
// @Summary      Create new customer
// @Description  Create a new customer record
// @Tags         customers
// @Accept       json
// @Produce      json
// @Param        customer  body      models.CreateCustomerRequest  true  "Customer data"
// @Success      201       {object}  models.Customer
// @Failure      400       {object}  map[string]string
// @Router       /customers [post]
// @Security     BearerAuth
func CreateCustomer(c *gin.Context) {
	var req models.CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var customer models.Customer
	err := db.PrimaryDB.QueryRow(
		"INSERT INTO customers (name, email) VALUES ($1, $2) RETURNING id, name, email, created_at, updated_at",
		req.Name, req.Email,
	).Scan(&customer.ID, &customer.Name, &customer.Email, &customer.CreatedAt, &customer.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create customer"})
		return
	}

	c.JSON(http.StatusCreated, customer)
}

// UpdateCustomer updates an existing customer
// @Summary      Update customer
// @Description  Update an existing customer record
// @Tags         customers
// @Accept       json
// @Produce      json
// @Param        id         path      int                           true  "Customer ID"
// @Param        customer   body      models.UpdateCustomerRequest  true  "Updated customer data"
// @Success      200        {object}  models.Customer
// @Failure      400        {object}  map[string]string
// @Failure      404        {object}  map[string]string
// @Router       /customers/{id} [put]
// @Security     BearerAuth
func UpdateCustomer(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	var req models.UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var customer models.Customer
	err = db.PrimaryDB.QueryRow(
		"UPDATE customers SET name = $1, email = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3 RETURNING id, name, email, created_at, updated_at",
		req.Name, req.Email, id,
	).Scan(&customer.ID, &customer.Name, &customer.Email, &customer.CreatedAt, &customer.UpdatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update customer"})
		return
	}

	c.JSON(http.StatusOK, customer)
}

// DeleteCustomer deletes a customer
// @Summary      Delete customer
// @Description  Delete a customer by ID
// @Tags         customers
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Customer ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /customers/{id} [delete]
// @Security     BearerAuth
func DeleteCustomer(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	result, err := db.PrimaryDB.Exec("DELETE FROM customers WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete customer"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer deleted successfully"})
}

