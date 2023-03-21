package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

type Short struct {
	Code      string
	Url       string
	Timestamp time.Time
	Ip        string
}

func Connect(path string) (err error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return
	}

	DB = db

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS shorts (code TEXT PRIMARY KEY, url TEXT NOT NULL, timestamp TEXT NOT NULL, ip TEXT NOT NULL)`)
	if err != nil {
		log.Println(err)
		func() {
			err := db.Close()
			log.Println(err)
		}()
	}

	return
}
