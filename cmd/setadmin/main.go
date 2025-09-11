package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./storage/sso.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("UPDATE users SET is_admin = true WHERE id = 1")
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Complite")
}
