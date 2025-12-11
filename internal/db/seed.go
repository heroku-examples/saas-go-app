package db

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

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
	
	// Check if we should generate performance demo data
	if os.Getenv("SEED_PERFORMANCE_DATA") == "true" {
		return SeedPerformanceData()
	}
	
	return SeedData()
}

// ClearAndReseed clears existing data and reseeds the database
// This is useful for regenerating demo data
func ClearAndReseed() error {
	log.Println("Clearing existing data...")
	
	// Clear accounts first (due to foreign key constraint)
	_, err := PrimaryDB.Exec("TRUNCATE TABLE accounts CASCADE")
	if err != nil {
		return fmt.Errorf("failed to clear accounts: %w", err)
	}
	
	// Clear customers
	_, err = PrimaryDB.Exec("TRUNCATE TABLE customers CASCADE")
	if err != nil {
		return fmt.Errorf("failed to clear customers: %w", err)
	}
	
	log.Println("Data cleared successfully")
	
	// Reseed based on environment variables
	if os.Getenv("SEED_PERFORMANCE_DATA") == "true" {
		return SeedPerformanceData()
	}
	
	return SeedData()
}

// SeedPerformanceData generates large datasets for NGPG performance demonstrations
// This creates thousands of customers and accounts to showcase:
// - Read scaling with follower pools
// - Analytics query performance
// - Automatic query routing
func SeedPerformanceData() error {
	log.Println("Generating performance demo data for NGPG showcase...")
	
	// Get configuration from environment or use defaults
	numCustomers := getEnvInt("SEED_CUSTOMERS", 1000)
	numAccountsPerCustomer := getEnvInt("SEED_ACCOUNTS_PER_CUSTOMER", 5)
	
	totalAccounts := numCustomers * numAccountsPerCustomer
	
	log.Printf("Generating %d customers with ~%d accounts each (~%d total accounts)...", 
		numCustomers, numAccountsPerCustomer, totalAccounts)
	
	// Create default test user if users table is empty
	var userCount int
	err := PrimaryDB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err == nil && userCount == 0 {
		passwordHash, err := auth.HashPassword("admin123")
		if err == nil {
			_, err = PrimaryDB.Exec(
				"INSERT INTO users (username, password_hash) VALUES ($1, $2)",
				"admin", passwordHash,
			)
			if err == nil {
				log.Println("Created default test user: username='admin', password='admin123'")
			}
		}
	}
	
	// Company name templates for realistic data
	companyTypes := []string{
		"Corporation", "Inc", "LLC", "Ltd", "Group", "Solutions", "Systems",
		"Innovations", "Technologies", "Enterprises", "Partners", "Associates",
		"Industries", "Holdings", "Ventures", "Capital", "Global", "International",
	}
	
	companyNames := []string{
		"Acme", "TechStart", "Global", "Digital", "Enterprise", "Premier", "Elite",
		"Advanced", "Strategic", "Dynamic", "Progressive", "Innovative", "Modern",
		"NextGen", "Future", "Vision", "Prime", "Apex", "Summit", "Peak",
		"Alpha", "Beta", "Gamma", "Delta", "Omega", "Nova", "Stellar", "Quantum",
		"Cyber", "Cloud", "Data", "Info", "Net", "Web", "Mobile", "Smart",
		"Fast", "Swift", "Rapid", "Turbo", "Power", "Force", "Strong", "Mighty",
	}
	
	accountTypes := []string{
		"Premium", "Enterprise", "Business", "Professional", "Standard", "Basic",
		"Starter", "Trial", "Pro", "Corporate", "Elite", "Ultimate", "Advanced",
		"Legacy", "Archive", "Development", "Production", "Staging", "Testing",
	}
	
	statuses := []string{"active", "inactive", "suspended", "pending"}
	statusWeights := []int{70, 20, 5, 5} // 70% active, 20% inactive, etc.
	
	rand.Seed(time.Now().UnixNano())
	
	// Generate customers - insert individually to get IDs
	customerIDs := make([]int, 0, numCustomers)
	
	log.Println("Creating customers...")
	startTime := time.Now()
	
	for i := 0; i < numCustomers; i++ {
		companyName := companyNames[rand.Intn(len(companyNames))]
		companyType := companyTypes[rand.Intn(len(companyTypes))]
		name := fmt.Sprintf("%s %s", companyName, companyType)
		email := fmt.Sprintf("contact@%s%d.com", 
			companyName[:min(len(companyName), 8)], 
			i)
		
		var id int
		err := PrimaryDB.QueryRow(
			"INSERT INTO customers (name, email) VALUES ($1, $2) RETURNING id",
			name, email,
		).Scan(&id)
		if err != nil {
			return fmt.Errorf("failed to insert customer: %w", err)
		}
		customerIDs = append(customerIDs, id)
		
		if (i+1)%100 == 0 {
			log.Printf("  Created %d/%d customers...", i+1, numCustomers)
		}
	}
	
	customerTime := time.Since(startTime)
	log.Printf("Created %d customers in %v", len(customerIDs), customerTime)
	
	// Generate accounts in batches for better performance
	log.Println("Creating accounts...")
	accountStartTime := time.Now()
	
	accountCount := 0
	accountBatch := make([]struct {
		customerID int
		name       string
		status     string
	}, 0, 500)
	
	insertAccountBatch := func() error {
		if len(accountBatch) == 0 {
			return nil
		}
		
		// Build batch insert query
		placeholders := ""
		values := make([]interface{}, 0, len(accountBatch)*3)
		
		for i, acc := range accountBatch {
			if i > 0 {
				placeholders += ", "
			}
			placeholders += fmt.Sprintf("($%d, $%d, $%d)", len(values)+1, len(values)+2, len(values)+3)
			values = append(values, acc.customerID, acc.name, acc.status)
		}
		
		query := fmt.Sprintf("INSERT INTO accounts (customer_id, name, status) VALUES %s", placeholders)
		_, err := PrimaryDB.Exec(query, values...)
		if err != nil {
			return fmt.Errorf("failed to insert account batch: %w", err)
		}
		
		accountCount += len(accountBatch)
		accountBatch = accountBatch[:0] // Clear batch
		return nil
	}
	
	for i := 0; i < len(customerIDs); i++ {
		customerID := customerIDs[i]
		accountsForCustomer := numAccountsPerCustomer
		
		// Add some variation: 20% of customers have 1-2x the average
		if rand.Float32() < 0.2 {
			accountsForCustomer = int(float32(accountsForCustomer) * (1.0 + rand.Float32()))
		}
		
		// Add accounts to batch
		for j := 0; j < accountsForCustomer; j++ {
			accountType := accountTypes[rand.Intn(len(accountTypes))]
			accountName := fmt.Sprintf("%s Account", accountType)
			status := weightedRandomStatus(statuses, statusWeights)
			
			accountBatch = append(accountBatch, struct {
				customerID int
				name       string
				status     string
			}{customerID, accountName, status})
			
			// Insert batch when it reaches size limit
			if len(accountBatch) >= 500 {
				if err := insertAccountBatch(); err != nil {
					return err
				}
			}
		}
		
		// Log progress
		if (i+1)%100 == 0 {
			log.Printf("  Created accounts for %d/%d customers (%d total accounts)...", 
				i+1, len(customerIDs), accountCount)
		}
	}
	
	// Insert remaining accounts
	if err := insertAccountBatch(); err != nil {
		return err
	}
	
	accountTime := time.Since(accountStartTime)
	totalTime := customerTime + accountTime
	
	log.Printf("Created %d accounts in %v", accountCount, accountTime)
	log.Printf("Performance demo data generation completed in %v", totalTime)
	log.Printf("Summary: %d customers, %d accounts", len(customerIDs), accountCount)
	
	return nil
}

// Helper functions
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("Warning: Invalid value for %s (%s), using default %d", key, value, defaultValue)
		return defaultValue
	}
	return intValue
}

func weightedRandomStatus(statuses []string, weights []int) string {
	totalWeight := 0
	for _, w := range weights {
		totalWeight += w
	}
	
	r := rand.Intn(totalWeight)
	cumulative := 0
	for i, weight := range weights {
		cumulative += weight
		if r < cumulative {
			return statuses[i]
		}
	}
	return statuses[len(statuses)-1]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}




