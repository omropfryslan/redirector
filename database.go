package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/net/idna"
)

// Database interface
type Database interface {
	Get(domain string) (string, bool, int, error)
	GetAll() ([]Domain, error)
	Save(url Domain) (int64, error)
	Delete(url Domain) (int64, error)
}

type sqlite struct {
	Path string
}

// Domain struct
type Domain struct {
	ID          int    `json:"id"`
	Domain      string `json:"domain"`
	Destination string `json:"destination"`
	Append      bool   `json:"append"`
	Code        int    `json:"code"`
}

func (s sqlite) Get(domain string) (string, bool, int, error) {
	db, err := sql.Open("sqlite3", s.Path)
	stmt, err := db.Prepare("select append, destination, code from sites where domain = ?")
	if err != nil {
		return "", false, 0, err
	}
	defer stmt.Close()
	var (
		destination string
		append      bool
		code        int
	)

	err = stmt.QueryRow(domain).Scan(&append, &destination, &code)

	if err != nil {
		return "", false, 0, err
	}

	return destination, append, code, nil
}

func (s sqlite) GetAll() ([]Domain, error) {
	db, err := sql.Open("sqlite3", s.Path)
	stmt, err := db.Query("select * from sites")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var domains []Domain

	for stmt.Next() {
		var row = &Domain{}

		err = stmt.Scan(&row.ID, &row.Domain, &row.Append, &row.Destination, &row.Code)
		row.Domain, _ = idna.ToUnicode(row.Domain)

		domains = append(domains, *row)
	}

	if err != nil {
		return nil, err
	}

	return domains, nil
}

func (s sqlite) Save(domain Domain) (int64, error) {
	db, err := sql.Open("sqlite3", s.Path)
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	stmt, err := tx.Prepare("replace into sites (id, domain, append, destination, code) values (?, ?, ?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var result sql.Result

	if domain.ID == 0 {
		result, err = stmt.Exec(nil, domain.Domain, domain.Append, domain.Destination, domain.Code)
	} else {
		result, err = stmt.Exec(domain.ID, domain.Domain, domain.Append, domain.Destination, domain.Code)
	}

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, nil
	}
	tx.Commit()

	return id, nil
}

func (s sqlite) Delete(domain Domain) (int64, error) {
	db, err := sql.Open("sqlite3", s.Path)
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	stmt, err := tx.Prepare("DELETE FROM sites where id = ?")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var result sql.Result

	result, err = stmt.Exec(domain.ID)

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, nil
	}
	tx.Commit()

	return id, nil
}

func (s sqlite) Init() {
	c, err := sql.Open("sqlite3", s.Path)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	sqlStmt := `CREATE TABLE IF NOT EXISTS sites (id INTEGER PRIMARY KEY AUTOINCREMENT, domain VARCHAR (255) NOT NULL UNIQUE, append BOOLEAN, destination VARCHAR (255) NOT NULL, code INTEGER DEFAULT (301) NOT NULL);
		CREATE INDEX IF NOT EXISTS domain ON sites (domain);`
	_, err = c.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}
}
