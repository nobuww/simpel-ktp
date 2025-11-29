set dotenv-load
set shell := ["sh.exe", "-c"]

# Run the full stack (Air + Tailwind + Bun Watch)
dev:
    @echo "Starting watchers and Air..."
    @trap 'kill 0' EXIT; \
    bunx @tailwindcss/cli -i ./assets/css/input.css -o ./static/css/style.css --watch & \
    bun build ./assets/js/main.js --outdir ./static/js --watch & \
    go tool air

# Production Build (Clean -> Generate -> Minify -> Compile)
build: clean
    go tool templ generate
    just build-assets
    go build -o ./bin/app ./cmd/server/main.go

build-backend:
    go tool templ generate
    go build -o ./tmp/main.exe ./cmd/server/main.go

# Watch CSS and JS for changes (Standalone)
watch-assets:
    bunx @tailwindcss/cli -i ./assets/css/input.css -o ./static/css/style.css --watch & \
    bun build ./assets/js/main.js --outdir ./static/js --watch & \
    wait

# One-off build for CSS and JS (Minified)
build-assets:
    bunx @tailwindcss/cli -i ./assets/css/input.css -o ./static/css/style.css --minify
    bun build ./assets/js/main.js --outdir ./static/js --minify

# Remove generated artifacts
clean:
    rm -rf ./static ./tmp ./bin

# Run Air hot-reload
air:
    go tool air

# Format all Go and Templ files
fmt:
    go tool templ fmt .
    go fmt ./...

# Generate Go code from SQL (sqlc)
sqlc:
    CGO_ENABLED=0 go tool sqlc generate

# Run database migrations up
migrate-up:
    go tool goose -dir db/migrations postgres "$DATABASE_URL" up

# Rollback the last migration
migrate-down:
    go tool goose -dir db/migrations postgres "$DATABASE_URL" down

# Create a new migration file (usage: just migrate-create add_users)
migrate-create name:
    go tool goose -dir db/migrations create {{name}} sql

# Check migration status
migrate-status:
    go tool goose -dir db/migrations postgres "$DATABASE_URL" status

# Reset all migrations
db-reset:
    go tool goose -dir db/migrations postgres "$DATABASE_URL" down-to 0
    just migrate-up