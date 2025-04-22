package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func initDB() {
	var err error

	db, err = sql.Open("sqlite3", "./chat.db")
	if err != nil {
		log.Fatal("Failed to connect to DB: %v", err)
	}

	createUserTable := `CREATE TABLE IF NOT EXISTS users(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT NOT NULL UNIQUE,
	password TEXT NOT NULL
	)`

	_, err = db.Exec(createUserTable)
	if err != nil {
		log.Fatal("Failed to create users table : ", err)
	}

	createMessageTable := `
	CREATE TABLE IF NOT EXISTS messages(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT NOT NULL,
	message TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP	)
	`

	_, err = db.Exec(createMessageTable)
	if err != nil {
		log.Fatal("Failed to create message table : %v", err)
	}
	
}
