// Package main provides a REST API server for managing truth or dare questions. The server uses a MySQL database to store questions and tags.
//
// @title Truth or Dare API
// @version 1.0
// @description A comprehensive REST API for managing and retrieving truth or dare questions. Supports filtering by language, type, and tags.
// @host localhost:8080
// @BasePath /api
// @schemes http
// @contact.name API Support
// @contact.url https://github.com/2Friendly4You/TruthOrDare
// @license.name MIT
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	_ "github.com/2Friendly4You/TruthOrDare/docs" // Generated swagger docs
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
)

// ErrorResponse represents a standard error response
// @Description Standard error response format
type ErrorResponse struct {
	// Error message describing what went wrong
	Message string `json:"message"`
	// Optional error code for client handling
	Code string `json:"code,omitempty"`
}

// Question represents a truth or dare question with its associated metadata
// @Description A truth or dare question entry with metadata
type Question struct {
	// Unique identifier for the question
	// @example 1
	ID int `json:"id"`

	// ISO language code of the question
	// @example "en"
	// @pattern ^[a-z]{2}$
	Language string `json:"language"`

	// Question type, either "truth" or "dare"
	// @example "truth"
	// @enum "truth" "dare"
	Type string `json:"type"`

	// The actual question or dare text
	// @example "What was your most embarrassing moment?"
	// @minLength 3
	Task string `json:"task"`

	// Array of associated tag names
	// @example ["funny","social","party"]
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

// @Summary Retrieve questions
// @Description Get a list of truth or dare questions with optional filtering capabilities
// @Tags questions
// @Accept json
// @Produce json
// @Param language query string false "ISO 639-1 language code filter (2 characters)" example(en)
// @Param type query string false "Question type filter" Enums(truth, dare)
// @Param tags query []string false "Filter questions by tags (comma-separated)" example(funny,party,social)
// @Param matchAllTags query boolean false "Require all specified tags to match (true) or any tag (false)" default(false)
// @Success 200 {array} Question "List of matching questions"
// @Failure 400 {object} ErrorResponse "Invalid request parameters"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /questions [get]
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

// @Summary Get available tags
// @Description Retrieve a list of all available tags that can be used for question filtering
// @Tags tags
// @Accept json
// @Produce json
// @Success 200 {array} string "List of available tags"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Example 200 {array} string ["funny", "social", "party", "deep", "romantic"]
// @Router /tags [get]
func getTags(w http.ResponseWriter, r *http.Request) {
	tags, err := db.GetTags()
	if err != nil {
		log.Printf("Failed to fetch tags: %v", err)
		http.Error(w, "Failed to fetch tags", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(tags); err != nil {
		log.Printf("Failed to encode tags to JSON: %v", err)
		http.Error(w, "Failed to encode tags to JSON", http.StatusInternalServerError)
	}
}

// main initializes and starts the HTTP server.
// The server provides the following endpoints:
//   - GET /api/questions: Retrieve questions with optional filters
//   - GET /api/tags: Retrieve all available tags
//
// Required environment variables:
//   - APP_PORT: Port number for the HTTP server
//   - All database-related environment variables (see NewDatabase docs)
func main() {
	initializeDatabase()
	defer db.Close()

	// Swagger documentation endpoint
	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	http.HandleFunc("/api/questions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getQuestions(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/tags", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getTags(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	port := os.Getenv("APP_PORT")
	log.Printf("API server running on port %s", port)
	log.Printf("Swagger documentation available at http://localhost:%s/swagger/index.html", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
