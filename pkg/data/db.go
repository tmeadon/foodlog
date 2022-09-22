package data

import (
	"errors"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var (
	DB                        *sqlx.DB
	ErrNotFound               error = fmt.Errorf("record not found")
	ErrUniqueConstraintFailed error = errors.New("unqiue constraint failed")
)

func Setup(dbPath string) {
	DB = sqlx.MustConnect("sqlite3", dbPath)
	if err := DB.Ping(); err != nil {
		log.Fatal(err)
	}
	DB.MustExec("PRAGMA foreign_keys = ON")
	migrate()
}
