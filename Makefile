.PHONY: setup run test clean help db-up db-down db-logs db-shell server migrate.up migrate.down make migrate.new debug integrations.test

help:
	@echo "Available commands:"
	@echo "  setup    - Set up the project environment and start database"
	@echo "  run      - Run the application"
	@echo "  test     - Run tests"
	@echo "  clean    - Clean up temporary files and stop containers"
	@echo "  db-up    - Start the PostgreSQL database"
	@echo "  db-down  - Stop the PostgreSQL database"
	@echo "  db-logs  - Show database logs"
	@echo "  db-shell - Connect to the database shell"
	@echo "  help     - Show this help message"

setup:
	@echo "Setting up the project..."
	@echo "Starting PostgreSQL database..."
	@docker-compose up -d
	@echo "Waiting for database to be ready..."
	@sleep 5
	@echo "Setup complete!"

run:
	@echo "Running the application..."
	@./bin/start.sh

test:
	@echo "Running tests..."
	@echo "Starting test database if not running..."
	@docker-compose up -d
	@echo "Running test script..."
	@./bin/test.sh

clean:
	@echo "Cleaning up..."
	@echo "Stopping and removing containers..."
	@docker-compose down -v
	@echo "Removing any temporary files..."
	@rm -rf *.log *.tmp

db-up:
	@echo "Starting PostgreSQL database..."
	@docker-compose up -d

db-down:
	@echo "Stopping PostgreSQL database..."
	@docker-compose down

db-logs:
	@echo "Showing database logs..."
	@docker-compose logs -f postgres

db-shell:
	@echo "Connecting to database shell..."
	@docker-compose exec postgres psql -U messaging_user -d messaging_service

server: migrate.up
	@echo "Starting the server..."
	@go run main.go serve

integrations.test:
	@echo "Running integration tests..."
	@echo "Starting test database if not running..."
	@docker-compose up -d
	@echo "Running test script..."
	@go test -v -count=1 ./internal/app/integrationtests

migrate.up:
	@echo "Running migrations..."
	@go run main.go migrate --direction up

migrate.down:
	@echo "Rolling back migrations..."
	@go run main.go migrate --direction down

migrate.new:
	@echo "Creating new migration files..."
	@migrate create -ext sql -dir migrations -seq $(migration_name)

debug:
	@echo "Running the application in debug mode..."
	@dlv debug  --listen=:2345 --headless=true --api-version=2 main.go -- serve 