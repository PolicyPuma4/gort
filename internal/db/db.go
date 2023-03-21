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

type Visit struct {
	Code      string
	Timestamp time.Time
	Ip        string
	UserAgent string
}

func Connect(path string) (err error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return
	}

	DB = db

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS shorts (code TEXT PRIMARY KEY, url TEXT NOT NULL, timestamp TEXT NOT NULL, ip TEXT NOT NULL)`)
	if err != nil {
		func() {
			err := db.Close()
			log.Println(err)
		}()

		return
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS visits (code TEXT NOT NULL, timestamp TEXT NOT NULL, ip TEXT NOT NULL, user_agent TEXT NOT NULL)`)
	if err != nil {
		func() {
			err := db.Close()
			log.Println(err)
		}()
	}

	return
}
