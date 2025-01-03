package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type Question struct {
	ID       int      `json:"id"`
	Language string   `json:"language"`
	Type     string   `json:"type"`
	Task     string   `json:"task"`
	Tags     []string `json:"tags"`
}

var db *sql.DB

func initializeDatabase() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_DATABASE"),
	)

	var dbErr error
	for i := 0; i < 10; i++ {
		db, dbErr = sql.Open("mysql", dsn)
		if dbErr == nil {
			dbErr = db.Ping()
			if dbErr == nil {
				break
			}
		}
		log.Printf("Failed to connect to database (attempt %d/10): %v", i+1, dbErr)
		time.Sleep(5 * time.Second)
	}

	if dbErr != nil {
		log.Fatalf("Failed to connect to database after 10 attempts: %v", dbErr)
	}

	log.Println("Connected to the database.")
}

func getQuestions(w http.ResponseWriter, r *http.Request) {
	language := r.URL.Query().Get("language")
	qType := r.URL.Query().Get("type")

	query := `
		SELECT DISTINCT q.id, q.language, q.type, q.task, GROUP_CONCAT(t.name) as tags
		FROM questions q
		LEFT JOIN question_tags qt ON q.id = qt.question_id
		LEFT JOIN tags t ON qt.tag_id = t.id
		WHERE (? = '' OR q.language = ?)
		AND (? = '' OR q.type = ?)
		GROUP BY q.id`

	rows, err := db.Query(query, language, language, qType, qType)
	if err != nil {
		log.Printf("Failed to fetch questions: %v", err)
		http.Error(w, "Failed to fetch questions", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var questions []Question
	for rows.Next() {
		var q Question
		var tags sql.NullString
		err := rows.Scan(&q.ID, &q.Language, &q.Type, &q.Task, &tags)
		if err != nil {
			log.Printf("Failed to parse question: %v", err)
			http.Error(w, "Failed to parse questions", http.StatusInternalServerError)
			return
		}
		if tags.Valid {
			q.Tags = strings.Split(tags.String, ",")
		} else {
			q.Tags = []string{}
		}
		questions = append(questions, q)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(questions); err != nil {
		log.Printf("Failed to encode questions to JSON: %v", err)
		http.Error(w, "Failed to encode questions to JSON", http.StatusInternalServerError)
	}
}

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
