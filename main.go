package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// Item represents a simple data structure
type Item struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
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

func getAllItems(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name, price FROM items")
	if err != nil {
		log.Printf("Failed to fetch items: %v", err)
		http.Error(w, "Failed to fetch items", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Name, &item.Price)
		if err != nil {
			log.Printf("Failed to parse item: %v", err)
			http.Error(w, "Failed to parse items", http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
		http.Error(w, "Error iterating rows", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		log.Printf("Failed to encode items to JSON: %v", err)
		http.Error(w, "Failed to encode items to JSON", http.StatusInternalServerError)
	}
}

func main() {
	initializeDatabase()
	defer db.Close()

	http.HandleFunc("/api/items", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getAllItems(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	port := os.Getenv("APP_PORT")
	log.Printf("API server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
