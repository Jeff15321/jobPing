package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// Test database connection
	dbURL := "postgres://jobscanner:password@localhost:5433/jobscanner?sslmode=disable"
	
	fmt.Printf("Attempting to connect to: %s\n", dbURL)
	
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	fmt.Println("Database opened successfully")

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Database ping successful!")
	
	// Test a simple query
	var version string
	err = db.QueryRow("SELECT version()").Scan(&version)
	if err != nil {
		log.Fatalf("Failed to query database: %v", err)
	}
	
	fmt.Printf("PostgreSQL version: %s\n", version)
}