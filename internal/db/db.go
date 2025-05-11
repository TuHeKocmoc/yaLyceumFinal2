package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var (
	GlobalDB *sql.DB
)

func InitDB() error {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "storage.db"
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("cannot open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("cannot ping db: %w", err)
	}
	GlobalDB = db

	if err := createTables(db); err != nil {
		return fmt.Errorf("cannot create tables: %w", err)
	}

	log.Println("[DB] SQLite initialized at", dbPath)
	return nil
}

func createTables(db *sql.DB) error {
	usersTable := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        login TEXT NOT NULL UNIQUE,
        password_hash TEXT NOT NULL
    );
    `

	expressionsTable := `
    CREATE TABLE IF NOT EXISTS expressions (
        id TEXT PRIMARY KEY,
        user_id INTEGER NOT NULL,
        raw TEXT NOT NULL,
        status TEXT NOT NULL,
        result REAL,
        final_task_id INTEGER,
        FOREIGN KEY(user_id) REFERENCES users(id)
    );
    `

	tasksTable := `
    CREATE TABLE IF NOT EXISTS tasks (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        expression_id TEXT NOT NULL,
        op TEXT NOT NULL,
        arg1_value REAL,
        arg1_task_id INTEGER,
        arg2_value REAL,
        arg2_task_id INTEGER,
        result REAL,
        status TEXT NOT NULL,
        FOREIGN KEY(expression_id) REFERENCES expressions(id)
    );
    `

	if _, err := db.Exec(usersTable); err != nil {
		return err
	}
	if _, err := db.Exec(expressionsTable); err != nil {
		return err
	}
	if _, err := db.Exec(tasksTable); err != nil {
		return err
	}

	return nil
}
