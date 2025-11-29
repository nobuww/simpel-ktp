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
  powershell -c "irm bun.sh/install.ps1 | iex"
  ```
* **Just** for convenience of managing tasks.
  ```powershell
  # install using bun/npm/pnpm/yarn
  bun install -g rust-just # or npm install -g rust-just

  # install using winget
  winget install --id Casey.Just --exact
  ```
  See [here](https://github.com/casey/just?tab=readme-ov-file#installation) for other installation instructions.

---

## Quick start
### 1) Install dependencies
```bash
bun install
```

> **Before running the development server, download Go modules:**
> ```bash
> go mod download
> ```

### 2) Development - It uses a Justfile with a few convenient tasks
The repository includes a `Justfile` with predefined workflows. If you have the `just` CLI installed, run:
```bash
just dev
```
This runs the Tailwind watcher, the Bun JS watcher, and a hot-reloading Go server via Air.

### 3) Build
To produce minified assets and build a binary (this matches the `build` task in the Justfile):
```bash
just build
```
Adjust the binary name and path for your target OS.

---

## Useful commands
* `just watch-assets` - watch and rebuild CSS/JS when sources change.
* `just air` - run Air hot-reload for the Go server.
* `just dev` - run all development watchers and the Go server (same as manual Taskfile `dev`).
* `just build` - generate templates, minify assets, and build the binary.
* `just fmt` - format Go and templ files.
* `go run ./cmd/main.go` - run the server without hot reload.

---

## Notes & Technical details
* The server binds to port `:8080` by default (see `cmd/main.go`).
* Static assets are served from `/assets` using the `assets` directory.
* Templates are generated using `go tool templ generate` (see the `build` task in the Justfile).
* Tailwind is invoked via `bunx` which runs CLI packages installed by `bun` locally.

---

## Troubleshooting
* If the CSS doesn’t update, ensure the Tailwind watcher is running and writing to `assets/css/output.css`.
* If Go build fails with missing modules, run `go mod tidy` to tidy and install dependencies.
* On Windows, confirm `bun` and `just` are available in your shell PATH.

---

## Development hints
* Run `just fmt` periodically to keep files formatted.
* Use the Justfile for common commands; it runs the recommended flags and sequences used by the project.