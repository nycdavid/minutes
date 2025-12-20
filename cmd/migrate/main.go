package main

import (
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	// test/dev/prod
	env := os.Getenv("ENV")
	if env == "" {
		panic("env variable not set")
	}

	direction := os.Args[1]
	fmt.Println(direction)

	fmt.Println(fmt.Sprintf("[%s] Running migrations...", env))

	m, err := migrate.New(
		"file://db/migrations",
		fmt.Sprintf("sqlite3://db.%s.sqlite", env),
	)

	if err != nil {
		panic(err)
	}

	if direction == "down" {
		if err := m.Down(); err != nil {
			panic(err)
		}
	} else {
		if err := m.Up(); err != nil {
			panic(err)
		}
	}
}
