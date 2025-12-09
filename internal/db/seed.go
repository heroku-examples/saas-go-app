package db

import (
	"database/sql"
	"log"

	"saas-go-app/internal/auth"
)

// SeedData populates the database with sample customers and accounts
func SeedData() error {
	// Check if data already exists
	var count int
	err := PrimaryDB.QueryRow("SELECT COUNT(*) FROM customers").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		log.Println("Database already contains data, skipping seed")
		return nil
	}

	log.Println("Seeding database with sample data...")

	// Create default test user if users table is empty
	var userCount int
	err = PrimaryDB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err == nil && userCount == 0 {
		// Create default test user: admin / admin123
		passwordHash, err := auth.HashPassword("admin123")
		if err == nil {
			_, err = PrimaryDB.Exec(
				"INSERT INTO users (username, password_hash) VALUES ($1, $2)",
				"admin", passwordHash,
			)
			if err == nil {
				log.Println("Created default test user: username='admin', password='admin123'")
			} else {
				log.Printf("Warning: Failed to create default user: %v", err)
			}
		} else {
			log.Printf("Warning: Failed to hash password for default user: %v", err)
		}
	}

	// Sample customers
	customers := []struct {
		name  string
		email string
	}{
		{"Acme Corporation", "contact@acme.com"},
		{"TechStart Inc", "info@techstart.com"},
		{"Global Solutions Ltd", "hello@globalsolutions.com"},
		{"Digital Innovations", "support@digitalinnovations.com"},
		{"Enterprise Systems", "sales@enterprisesystems.com"},
	}

	customerIDs := make([]int, 0, len(customers))

	// Insert customers
	for _, customer := range customers {
		var id int
		err := PrimaryDB.QueryRow(
			"INSERT INTO customers (name, email) VALUES ($1, $2) RETURNING id",
			customer.name, customer.email,
		).Scan(&id)
		if err != nil {
			return err
		}
		customerIDs = append(customerIDs, id)
		log.Printf("Created customer: %s (ID: %d)", customer.name, id)
	}

	// Sample accounts linked to customers
	accounts := []struct {
		customerIndex int // Index into customerIDs array
		name          string
		status        string
	}{
		// Acme Corporation (index 0)
		{0, "Premium Account", "active"},
		{0, "Basic Account", "active"},
		{0, "Trial Account", "inactive"},
		// TechStart Inc (index 1)
		{1, "Enterprise Account", "active"},
		{1, "Starter Account", "active"},
		// Global Solutions Ltd (index 2)
		{2, "Corporate Account", "active"},
		{2, "Legacy Account", "inactive"},
		// Digital Innovations (index 3)
		{3, "Pro Account", "active"},
		// Enterprise Systems (index 4)
		{4, "Business Account", "active"},
		{4, "Standard Account", "active"},
		{4, "Archive Account", "inactive"},
	}

	// Insert accounts
	for _, account := range accounts {
		customerID := customerIDs[account.customerIndex]
		var id int
		err := PrimaryDB.QueryRow(
			"INSERT INTO accounts (customer_id, name, status) VALUES ($1, $2, $3) RETURNING id",
			customerID, account.name, account.status,
		).Scan(&id)
		if err != nil {
			return err
		}
		log.Printf("Created account: %s (ID: %d) for customer ID: %d", account.name, id, customerID)
	}

	log.Println("Database seeding completed successfully")
	return nil
}

// SeedDataIfEmpty seeds data only if the database is empty
func SeedDataIfEmpty() error {
	var count int
	err := PrimaryDB.QueryRow("SELECT COUNT(*) FROM customers").Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if count > 0 {
		return nil // Database already has data
	}
	return SeedData()
}




