package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func Run() {
	cmd := &cli.Command{
		Commands: []*cli.Command{
			{
				Name:    "migrate",
				Aliases: []string{"m"},
				Usage:   "Run database migrations",
				Action: func(context.Context, *cli.Command) error {
					fmt.Println("Running migrations...")
					db, err := sql.Open("postgres", "postgres://messaging_user:messaging_password@localhost:5432/messaging_service?sslmode=disable")
					if err != nil {
						return fmt.Errorf("failed to connect to database: %w", err)
					}

					driver, err := postgres.WithInstance(db, &postgres.Config{})
					if err != nil {
						return fmt.Errorf("failed to create database driver: %w", err)
					}

					m, err := migrate.NewWithDatabaseInstance(
						"file://migrations",
						"postgres", driver)
					if err != nil {
						return fmt.Errorf("failed to create migration instance: %w", err)
					}

					if err := m.Up(); err != nil {
						log.Fatal(err)
						return fmt.Errorf("failed to apply migrations: %w", err)

					}
					fmt.Println("Migrations applied successfully")
					return nil
				},
			},
			{
				Name:    "server",
				Aliases: []string{"s"},
				Usage:   "Start the server",
				Action: func(context.Context, *cli.Command) error {
					fmt.Println("Starting server...")
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
