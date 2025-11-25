# Simpel KTP

## Project Overview
This repository contains a Go server that serves HTML templates and static assets. It uses:
* **Go** for the server and templates.
* **Chi** as the HTTP router.
* **Tailwind CSS** for styling via the Tailwind CLI (executed using bun).
* **Air** for rapid Go hot reload during development.
* **Bun** to run Node-style tooling (tailwind CLI).

---

## Prerequisites
* **Go** (1.19+) installed and available on your PATH. Install from [https://go.dev/dl/](https://go.dev/dl/).
* **Bun** to run Tailwind CLI via `bunx`. See [here](https://bun.sh/) for installation instructions. On Windows PowerShell you can run:
  ```powershell
  iwr https://bun.sh/install -useb | iex
  ```
* **Task** (optional) if you want to use the Taskfile in the root for convenience. [Install here](https://taskfile.dev/installation/) or run commands manually listed below.

---

## Quick start
### 1) Install dependencies
```bash
bun install
```

### 2) Development - It uses a Taskfile with a few convenient tasks
The repository includes a Taskfile.yml with predefined workflows. If you have the `task` CLI installed, run:
```bash
task dev
```
This runs the Tailwind watcher and a hot-reloading Go server using Air.

### 2a) Manual development alternative
If you don’t have `task`, open two terminals and run the following commands separately:
1. Build Tailwind in watch mode (requires `bun`):
   ```bash
   bunx @tailwindcss/cli -i ./assets/css/input.css -o ./assets/css/output.css --watch
   ```
2. Run the Go server with automatic reload using Air (optional):
   ```bash
   go run github.com/air-verse/air@latest
   ```
   Or run the server directly (no auto-reload):
   ```bash
   go run ./cmd/main.go
   ```

### 3) Build
To produce a minified CSS and build a binary (this matches the Taskfile `build` task):
```bash
go tool templ generate
bunx @tailwindcss/cli -i ./assets/css/input.css -o ./assets/css/output.css --minify
go build -o ./tmp/main.exe ./cmd/main.go
```
Adjust the binary name and path for your target OS.

---

## Useful commands
* `task css` - watch and build CSS with Tailwind.
* `task air` - run Air hot-reload for Go server.
* `task dev` - run both CSS and Air for development.
* `task build` - generate templates, minify CSS and build binary.
* `task fmt` - format Go and templ files.
* `go run ./cmd/main.go` - run the server without hot reload.

---

## Notes & Technical details
* The server binds to port `:8080` by default (see `cmd/main.go`).
* Static assets are served from `/assets` using the `assets` directory.
* Templates are generated using `go tool templ generate` (see Taskfile).
* Tailwind is invoked via `bunx` which runs CLI packages installed by `bun` locally.

---

## Troubleshooting
* If the CSS doesn’t update, ensure the Tailwind watcher is running and writing to `assets/css/output.css`.
* If Go build fails with missing modules, run `go mod tidy` to tidy and install dependencies.
* On Windows, confirm `bun` and `task` are available in your shell PATH.

---

## Development hints
* Run `task fmt` periodically to keep files formatted.
* Use the Taskfile for common commands; it runs the recommended flags and sequence used by the project.