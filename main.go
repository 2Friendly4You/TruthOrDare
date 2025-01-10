// Package main provides a REST API server for managing truth or dare questions.
//
// @title Truth or Dare API
// @version 1.0
// @description A REST API for managing truth or dare questions in a MySQL database
// @host localhost:8080
// @BasePath /api
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

// @Summary Get questions
// @Description Retrieve questions with optional filters
// @Tags questions
// @Accept json
// @Produce json
// @Param language query string false "Filter by ISO language code (e.g., en, de)"
// @Param type query string false "Filter by question type (truth/dare)"
// @Param tags query []string false "Filter by multiple tags"
// @Param matchAllTags query boolean false "If true, all specified tags must match"
// @Success 200 {array} Question
// @Failure 500 {object} map[string]string
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

// @Summary Get all tags
// @Description Retrieve all available tags
// @Tags tags
// @Accept json
// @Produce json
// @Success 200 {array} string
// @Failure 500 {object} map[string]string
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
