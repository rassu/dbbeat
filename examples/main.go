package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/rassu/dbbeat/config"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("go run main.go <create/insert>")
	}
	mod := os.Args[1]
	c := config.DefaultConfig

	log.Printf("config %s", c.DBConfig.URI)
	db, err := sql.Open("postgres", c.DBConfig.URI)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if mod == "create" {
		err = create(db)
		if err != nil {
			log.Fatal(err)
		}
	} else if mod == "insert" {
		tables := []string{"person", "car", "sample"}
		for _, t := range tables {
			err = insert(db, fmt.Sprintf("%s.sql", t))
			if err != nil {
				log.Println(err)
			}
		}
	} else {
		log.Fatal("go run main.go <create/insert>")
	}
}

func create(db *sql.DB) error {
	f, err := os.Open("table.sql")
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if err != nil && err != io.EOF {
			return err
		}
		_, err = db.Exec(scanner.Text())
		if err != nil {
			return err
		}
	}
	return nil
}

func insert(db *sql.DB, tbl string) error {
	f, err := os.Open(tbl)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if err != nil && err != io.EOF {
			return err
		}
		_, err = db.Exec(scanner.Text())
		if err != nil {
			return err
		}
	}
	return nil
}
