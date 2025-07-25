package cmd

import (
	"context"
	"errors"
	"fmt"
	"hatchapp/config"
	"hatchapp/internal/app/server"
	"hatchapp/internal/pkg/migration"
	"hatchapp/internal/pkg/repository"
	"os"

	"github.com/labstack/gommon/log"

	"github.com/urfave/cli/v3"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

var dbFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  "direction",
		Value: "up",
		Usage: "Migration direction (up/down)",
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
}

var serviceFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "sendgrid-api-key",
		Value:   "your_sendgrid_api_key",
		Usage:   "SendGrid API Key",
		Sources: cli.EnvVars("SENDGRID_API_KEY"),
	},
	&cli.StringFlag{
		Name:    "sendgrid-account-sid",
		Value:   "your_twilio_account_sid",
		Usage:   "Twilio Account SID",
		Sources: cli.EnvVars("TWILIO_ACCOUNT_SID"),
	},
	&cli.StringFlag{
		Name:    "twilio-api-key",
		Value:   "your_twilio_api_key",
		Usage:   "Twilio API Key",
		Sources: cli.EnvVars("TWILIO_API_KEY"),
	},
	&cli.StringFlag{
		Name:    "twilio-account-sid",
		Value:   "your_twilio_account_sid",
		Usage:   "Twilio Account SID",
		Sources: cli.EnvVars("TWILIO_ACCOUNT_SID"),
	},
}

func Run() {
	cmd := &cli.Command{
		Commands: []*cli.Command{
			{
				Name:    "migrate",
				Aliases: []string{"m"},
				Flags:   dbFlags,
				Usage:   "Run database migrations",
				Action: func(ctx context.Context, cliCmd *cli.Command) error {
					log.Info("Preparing migrations...")

					host := cliCmd.String("db-host")
					port := cliCmd.String("db-port")
					username := cliCmd.String("db-username")
					password := cliCmd.String("db-password")
					dbName := cliCmd.String("db-name")

					connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", username, password, host, port, dbName)
					m, err := migration.Initialize(connectionString)
					if err != nil {
						return fmt.Errorf("failed to initialize migration: %w", err)
					}

					switch cliCmd.String("direction") {
					case "up":
						if err := m.Up(); err != nil {
							if err == migrate.ErrNoChange {
								log.Info("No migrations to apply")
								return nil
							}

							fmt.Printf("Type of error: %T\n", err)

							// Handle dirty state
							var dbErr database.Error
							var dirtyErr migrate.ErrDirty
							if errors.As(err, &dbErr) || errors.As(err, &dirtyErr) {
								err = migration.Rollback(m, "up")
								if err != nil {
									log.Errorf("failed to roll back migration: %v", err)
								}

								return nil
							}

							return fmt.Errorf("failed to apply migrations: %w", err)
						}
					case "down":
						if err := m.Down(); err != nil {
							if err == migrate.ErrNoChange {
								fmt.Println("No migrations to roll back")
								return nil
							}

							// Handle dirty state
							var dbErr database.Error
							var dirtyErr migrate.ErrDirty
							if errors.As(err, &dbErr) || errors.As(err, &dirtyErr) {
								err = migration.Rollback(m, "down")
								if err != nil {
									log.Errorf("failed to roll back migration: %v", err)
								}
							}

							return fmt.Errorf("failed to apply migrations: %w", err)
						}
					default:
						return fmt.Errorf("invalid migration direction: %s", cliCmd.String("direction"))
					}

					log.Info("migrations applied successfully")
					return nil
				},
			},
			{
				Name:    "serve",
				Aliases: []string{"s"},
				Usage:   "Start the server",
				Flags:   append(dbFlags, serviceFlags...),
				Action: func(ctx context.Context, cliCmd *cli.Command) error {
					log.Info("Starting server...")

					host := cliCmd.String("db-host")
					port := cliCmd.String("db-port")
					username := cliCmd.String("db-username")
					password := cliCmd.String("db-password")
					dbName := cliCmd.String("db-name")

					connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", username, password, host, port, dbName)

					twilioAccountSID := cliCmd.String("twilio-account-sid")
					twilioAPIKey := cliCmd.String("twilio-api-key")
					sendgridAPIKey := cliCmd.String("sendgrid-api-key")
					sendgridAccountSID := cliCmd.String("sendgrid-account-sid")

					appConfig := map[string]string{
						"db_connection_string": connectionString,
						"twilio_account_sid":   twilioAccountSID,
						"twilio_api_key":       twilioAPIKey,
						"sendgrid_api_key":     sendgridAPIKey,
						"sendgrid_account_sid": sendgridAccountSID,
					}
					ctx = config.SaveConfigToContext(ctx, appConfig)

					err := repository.Initialize(ctx)
					if err != nil {
						return fmt.Errorf("failed to initialize repository: %w", err)
					}

					err = server.Run(ctx)
					if err != nil {
						return fmt.Errorf("failed to start server: %w", err)
					}

					log.Info("Server has shut down gracefully")
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
