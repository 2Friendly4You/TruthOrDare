package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Database represents a connection to the MySQL database and provides
// methods for querying and manipulating truth or dare questions.
type Database struct {
	db *sql.DB
}

// QueryConfig contains configuration options for database queries.
type QueryConfig struct {
	// MatchAllTags determines whether all specified tags must match (true)
	// or if matching any tag is sufficient (false)
	MatchAllTags bool
}

// NewDatabase creates a new database connection using environment variables
// and returns a Database instance. It attempts to connect up to 10 times
// with a 5-second delay between attempts.
//
// Required environment variables:
//   - MYSQL_USER: Database username
//   - MYSQL_PASSWORD: Database password
//   - MYSQL_HOST: Database host address
//   - MYSQL_PORT: Database port
//   - MYSQL_DATABASE: Database name
//
// Returns:
//   - (*Database, nil) on successful connection
//   - (nil, error) if connection fails after 10 attempts
func NewDatabase() (*Database, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_DATABASE"),
	)

	var db *sql.DB
	var err error
	for i := 0; i < 10; i++ {
		db, err = sql.Open("mysql", dsn)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		log.Printf("Failed to connect to database (attempt %d/10): %v", i+1, err)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after 10 attempts: %v", err)
	}

	return &Database{db: db}, nil
}

// Close closes the database connection.
// It should be deferred after creating a new Database instance.
func (d *Database) Close() error {
	return d.db.Close()
}

// GetQuestions retrieves questions from the database based on specified filters.
//
// Parameters:
//   - language: Filter by language code (e.g., "en", "de"). Empty string matches all languages
//   - qType: Filter by question type ("truth" or "dare"). Empty string matches all types
//   - tags: Array of tag names to filter by. Empty array matches all tags
//   - config: Query configuration options. If nil, defaults to matching any tag
//
// Examples:
//
//	// Get all English questions
//	db.GetQuestions("en", "", nil, nil)
//
//	// Get German truth questions with either "18+" or "food" tags
//	db.GetQuestions("de", "truth", []string{"18+", "food"}, &QueryConfig{MatchAllTags: false})
//
//	// Get questions that have both "18+" AND "alcohol" tags
//	db.GetQuestions("", "", []string{"18+", "alcohol"}, &QueryConfig{MatchAllTags: true})
//
// Returns:
//   - ([]Question, nil) on success
//   - (nil, error) if the query fails
func (d *Database) GetQuestions(language, qType string, tags []string, config *QueryConfig) ([]Question, error) {
	baseQuery := `
		SELECT DISTINCT q.id, q.language, q.type, q.task, GROUP_CONCAT(t.name) as tags
		FROM questions q
		LEFT JOIN question_tags qt ON q.id = qt.question_id
		LEFT JOIN tags t ON qt.tag_id = t.id`

	whereConditions := []string{}
	args := []interface{}{}

	if language != "" {
		whereConditions = append(whereConditions, "q.language = ?")
		args = append(args, language)
	}

	if qType != "" {
		whereConditions = append(whereConditions, "q.type = ?")
		args = append(args, qType)
	}

	if len(tags) > 0 {
		if config != nil && config.MatchAllTags {
			// Match all tags using COUNT and HAVING
			baseQuery += fmt.Sprintf(`
				INNER JOIN (
					SELECT qt.question_id
					FROM question_tags qt
					INNER JOIN tags t ON qt.tag_id = t.id
					WHERE t.name IN (?%s)
					GROUP BY qt.question_id
					HAVING COUNT(DISTINCT t.name) = ?
				) matching_tags ON q.id = matching_tags.question_id`,
				strings.Repeat(",?", len(tags)-1))

			for _, tag := range tags {
				args = append(args, tag)
			}
			args = append(args, len(tags))
		} else {
			// Match any tag
			whereConditions = append(whereConditions, fmt.Sprintf("t.name IN (?%s)", strings.Repeat(",?", len(tags)-1)))
			for _, tag := range tags {
				args = append(args, tag)
			}
		}
	}

	if len(whereConditions) > 0 {
		baseQuery += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	baseQuery += " GROUP BY q.id"

	rows, err := d.db.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch questions: %v", err)
	}
	defer rows.Close()

	var questions []Question
	for rows.Next() {
		var q Question
		var tags sql.NullString
		err := rows.Scan(&q.ID, &q.Language, &q.Type, &q.Task, &tags)
		if err != nil {
			return nil, fmt.Errorf("failed to parse question: %v", err)
		}
		if tags.Valid {
			q.Tags = strings.Split(tags.String, ",")
		} else {
			q.Tags = []string{}
		}
		questions = append(questions, q)
	}

	return questions, nil
}

// GetTags returns all available tags in the database.
//
// Returns:
//   - ([]string, nil) containing all tag names on success
//   - (nil, error) if the query fails
func (d *Database) GetTags() ([]string, error) {
	rows, err := d.db.Query("SELECT name FROM tags")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tags: %v", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		err := rows.Scan(&tag)
		if err != nil {
			return nil, fmt.Errorf("failed to parse tag: %v", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// AddQuestion adds a new question to the database along with its tags.
// If any of the specified tags don't exist, they will be created automatically.
// The operation is performed in a transaction to ensure data consistency.
//
// Parameters:
//   - q: Question struct containing:
//   - Language: Language code (required)
//   - Type: "truth" or "dare" (required)
//   - Task: The actual question/dare text (required)
//   - Tags: Array of tag names (optional)
//
// Example:
//
//	err := db.AddQuestion(Question{
//	    Language: "en",
//	    Type:    "truth",
//	    Task:    "What's your biggest fear?",
//	    Tags:    []string{"deep", "emotional"},
//	})
//
// Returns:
//   - nil on success
//   - error if the operation fails (transaction will be rolled back)
func (d *Database) AddQuestion(q Question) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	result, err := tx.Exec("INSERT INTO questions (language, type, task) VALUES (?, ?, ?)",
		q.Language, q.Type, q.Task)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to insert question: %v", err)
	}

	questionID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get last insert ID: %v", err)
	}

	for _, tag := range q.Tags {
		var tagID int64
		err := tx.QueryRow("SELECT id FROM tags WHERE name = ?", tag).Scan(&tagID)
		if err == sql.ErrNoRows {
			result, err := tx.Exec("INSERT INTO tags (name) VALUES (?)", tag)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to insert tag: %v", err)
			}
			tagID, err = result.LastInsertId()
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to get tag ID: %v", err)
			}
		} else if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to query tag: %v", err)
		}

		_, err = tx.Exec("INSERT INTO question_tags (question_id, tag_id) VALUES (?, ?)",
			questionID, tagID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert question tag: %v", err)
		}
	}

	return tx.Commit()
}
