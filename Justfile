set dotenv-load
set shell := ["sh.exe", "-c"]

# Run the full stack (Air + Vite Dev Server)
dev:
    @echo "Starting Vite and Air..."
    @trap 'kill 0' EXIT; \
    bun run dev & \
    go tool air

# Production Build (Clean -> Templ -> Vite Build -> Go Build)
build: clean
    go tool templ generate
    just build-assets
    go build -o ./bin/app.exe ./cmd/server/main.go

# build backend (used in .air.toml)
build-backend:
    go tool templ generate
    go build -o ./tmp/main.exe ./cmd/server/main.go

watch-assets:
    bun run dev

build-assets:
    bun run build

# Remove generated artifacts
clean:
    rm -rf ./static ./tmp ./bin ./dist

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

# apply migrations to db
migrate-up:
    go tool goose -dir db/migrations postgres "$DATABASE_URL" up

# cancel applied migrations
migrate-down:
    go tool goose -dir db/migrations postgres "$DATABASE_URL" down

# create new migrations file 
migrate-create name:
    go tool goose -dir db/migrations create {{name}} sql

# check migrations status
migrate-status:
    go tool goose -dir db/migrations postgres "$DATABASE_URL" status

# reset db
db-reset:
    go tool goose -dir db/migrations postgres "$DATABASE_URL" down-to 0
    just migrate-up