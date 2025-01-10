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

// Database represents a connection to the MySQL database
// @Description Database connection handler for truth or dare questions
type Database struct {
	db *sql.DB
}

// QueryConfig contains configuration options for database queries
// @Description Configuration options for filtering questions
type QueryConfig struct {
	// Determines if all tags must match (true) or any tag matches (false)
	// @example false
	MatchAllTags bool
}

// NewDatabase creates a new database connection using environment variables
// @Description Establishes database connection with retry mechanism
// @Return (*Database) Database connection instance
// @Return (error) Connection error if all attempts fail
// @x-envVars MYSQL_USER - Database username
// @x-envVars MYSQL_PASSWORD - Database password
// @x-envVars MYSQL_HOST - Database host address
// @x-envVars MYSQL_PORT - Database port number
// @x-envVars MYSQL_DATABASE - Database name
func NewDatabase() (*Database, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
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

// GetQuestions retrieves filtered questions from the database
// @Description Fetches questions based on language, type, and tags
// @Param language string ISO language code filter (e.g., "en", "de")
// @Param qType string Question type filter ("truth" or "dare")
// @Param tags []string Tag names to filter by
// @Param config *QueryConfig Query configuration options
// @Return []Question List of matching questions
// @Return error Query execution error
// @Example
//
//	// Get all English questions
//	questions, err := db.GetQuestions("en", "", nil, nil)
//
//	// Get German truth questions with specific tags
//	questions, err := db.GetQuestions("de", "truth", []string{"funny"}, &QueryConfig{MatchAllTags: true})
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

// GetTags returns all available question tags
// @Description Retrieves complete list of available tags from database
// @Return []string List of tag names
// @Return error Query execution error
// @Example
//
//	tags, err := db.GetTags()
//	// Returns: ["funny", "social", "party", "deep", "romantic"]
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

// AddQuestion inserts a new question with associated tags
// @Description Creates a new question and its tag associations in a transaction
// @Param q Question Question object containing all required fields
// @Return error Operation error if transaction fails
// @Example
//
//	err := db.AddQuestion(Question{
//	    Language: "en",
//	    Type:    "truth",
//	    Task:    "What's your biggest fear?",
//	    Tags:    []string{"deep", "emotional"},
//	})
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

// Close terminates the database connection
// @Description Safely closes the database connection and frees resources
// @Return error Connection closure error
func (d *Database) Close() error {
	return d.db.Close()
}
