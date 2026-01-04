package main

import (
	"errors"
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
	dbURL := os.Args[2]

	if dbURL == "" {
		fmt.Println("db url not set, skipping...")
		return
	}

	fmt.Print(fmt.Sprintf("[%s] Running migrations.... ", env))

	m, err := migrate.New(
		"file://db/migrations",
		fmt.Sprintf("sqlite3://%s", dbURL),
	)

	if err != nil {
		panic(err)
	}

	if direction == "down" {
		if err := m.Down(); err != nil {
			panic(err)
		}
	} else if direction == "up" {
		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("no changes")
			} else {
				panic(err)
			}
		}
	} else {
		panic("invalid direction")
	}
}
