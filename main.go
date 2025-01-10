// Package main provides a REST API server for managing truth or dare questions.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

// Question represents a truth or dare question with its associated metadata.
type Question struct {
	// ID is the unique identifier for the question
	ID int `json:"id"`

	// Language is the ISO language code (e.g., "en", "de")
	Language string `json:"language"`

	// Type must be either "truth" or "dare"
	Type string `json:"type"`

	// Task contains the actual question or dare text
	Task string `json:"task"`

	// Tags is an array of associated tag names
	Tags []string `json:"tags"`
}

var db *Database

// initializeDatabase loads environment variables and establishes
// the database connection. Exits the program if initialization fails.
func initializeDatabase() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	var dbErr error
	db, dbErr = NewDatabase()
	if dbErr != nil {
		log.Fatal(dbErr)
	}

	log.Println("Connected to the database.")
}

// getQuestions handles GET requests to /api/questions endpoint.
//
// Supported query parameters:
//   - language: Filter by language code (optional)
//   - type: Filter by "truth" or "dare" (optional)
//   - tags: Multiple tag filters (optional)
//   - matchAllTags: "true" to match all tags, "false" to match any (optional)
//
// Examples:
//
//	GET /api/questions?language=en
//	GET /api/questions?type=dare&tags=18+&tags=alcohol&matchAllTags=true
//
// With curl:
//
//	curl -X GET "http://localhost:<port>/api/questions?language=en&type=truth&tags=18%2B&tags=alcohol&matchAllTags=true"
//	curl -X GET "http://localhost:<port>/api/questions?language=de&type=dare&tags=18%2B&tags=food"
//	curl -X GET "http://localhost:<port>/api/questions?language=en"
//
// Response:
//
//	200 OK: JSON array of Question objects
//	500 Internal Server Error: If database query fails
func getQuestions(w http.ResponseWriter, r *http.Request) {
	language := r.URL.Query().Get("language")
	qType := r.URL.Query().Get("type")
	tags := r.URL.Query()["tags"]
	matchAllTags := r.URL.Query().Get("matchAllTags") == "true"

	config := &QueryConfig{
		MatchAllTags: matchAllTags,
	}

	// deepcode ignore Sqli: <is validated by the database driver>
	questions, err := db.GetQuestions(language, qType, tags, config)
	if err != nil {
		log.Printf("Failed to fetch questions: %v", err)
		http.Error(w, "Failed to fetch questions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(questions); err != nil {
		log.Printf("Failed to encode questions to JSON: %v", err)
		http.Error(w, "Failed to encode questions to JSON", http.StatusInternalServerError)
	}
}

// main initializes and starts the HTTP server.
// The server provides the following endpoints:
//   - GET /api/questions: Retrieve questions with optional filters
//
// Required environment variables:
//   - APP_PORT: Port number for the HTTP server
//   - All database-related environment variables (see NewDatabase docs)
func main() {
	initializeDatabase()
	defer db.Close()

	http.HandleFunc("/api/questions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getQuestions(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	port := os.Getenv("APP_PORT")
	log.Printf("API server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
