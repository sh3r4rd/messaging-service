package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hatchapp/internal/app/server"
	"hatchapp/internal/pkg/repository"
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
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "direction",
						Value: "up",
						Usage: "direction for the migration",
					},
					&cli.StringFlag{
						Name:    "db-username",
						Value:   "messaging_user",
						Usage:   "database username",
						Sources: cli.EnvVars("DB_USERNAME"),
					},
					&cli.StringFlag{
						Name:    "db-password",
						Value:   "messaging_password",
						Usage:   "database password",
						Sources: cli.EnvVars("DB_PASSWORD"),
					},
					&cli.StringFlag{
						Name:    "db-host",
						Value:   "localhost",
						Usage:   "database host",
						Sources: cli.EnvVars("DB_HOST"),
					},
					&cli.StringFlag{
						Name:    "db-port",
						Value:   "5432",
						Usage:   "database port",
						Sources: cli.EnvVars("DB_PORT"),
					},
					&cli.StringFlag{
						Name:    "db-name",
						Value:   "messaging_service",
						Usage:   "database name",
						Sources: cli.EnvVars("DB_NAME"),
					},
				},
				Usage: "Run database migrations",
				Action: func(ctx context.Context, cliCmd *cli.Command) error {
					fmt.Println("Preparing migrations...")

					host := cliCmd.String("db-host")
					port := cliCmd.String("db-port")
					username := cliCmd.String("db-username")
					password := cliCmd.String("db-password")
					dbName := cliCmd.String("db-name")

					connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", username, password, host, port, dbName)
					db, err := sql.Open("postgres", connectionString)
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

					switch cliCmd.String("direction") {
					case "up":
						if err := m.Up(); err != nil {
							if err == migrate.ErrNoChange {
								fmt.Println("No migrations to apply")
								return nil
							}

							// Handle dirty state
							var dirtyErr migrate.ErrDirty
							if errors.As(err, &dirtyErr) {
								version, _, verErr := m.Version()
								if verErr != nil {
									return fmt.Errorf("failed to get dirty migration version: %w", verErr)
								}
								fmt.Printf("Database is dirty at version %d. Forcing clean state...\n", version)
								if forceErr := m.Force(int(version)); forceErr != nil {
									return fmt.Errorf("failed to force migration version: %w", forceErr)
								}

								return fmt.Errorf("migration error, please re-run the migration after cleaning up")
							}
							return fmt.Errorf("failed to apply migrations: %w", err)
						}
					case "down":
						if err := m.Down(); err != nil {
							if err == migrate.ErrNoChange {
								fmt.Println("No migrations to roll back")
								return nil
							}

							return fmt.Errorf("failed to roll back migrations: %w", err)
						}
					default:
						return fmt.Errorf("invalid migration direction: %s", cliCmd.String("direction"))
					}

					fmt.Println("migrations applied successfully")
					return nil
				},
			},
			{
				Name:    "serve",
				Aliases: []string{"s"},
				Usage:   "Start the server",
				Action: func(context.Context, *cli.Command) error {
					fmt.Println("Starting server...")
					err := repository.Initialize()
					if err != nil {
						return fmt.Errorf("failed to initialize repository: %w", err)
					}

					err = server.Run()
					if err != nil {
						return fmt.Errorf("failed to start server: %w", err)
					}

					fmt.Println("Server has shut down gracefully")
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
