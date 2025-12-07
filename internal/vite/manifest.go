package vite

import (
	"encoding/json"
	"fmt"
	"os"
)

var manifest map[string]Chunk

type Chunk struct {
	File    string `json:"file"`
	Src     string `json:"src"`
	IsEntry bool   `json:"isEntry"`
}

var isDev bool

func Init(environment string) error {
	// Check if manifest exists to determine mode
	// If manifest.json exists, we assume we want to use the built assets (Production-like)
	// If it doesn't exist, we assume we are in Dev mode using the Vite server
	content, err := os.ReadFile("static/.vite/manifest.json")
	if err != nil {
		// Manifest not found, assume dev mode
		isDev = true
		return nil
	}

	// Manifest found, load it
	isDev = false
	if err := json.Unmarshal(content, &manifest); err != nil {
		return fmt.Errorf("could not parse vite manifest: %w", err)
	}

	return nil
}

func IsDev() bool {
	return isDev
}

func AssetPath(path string) string {
	if isDev {
		return fmt.Sprintf("http://localhost:5173/%s", path)
	}

	if chunk, ok := manifest[path]; ok {
		return "/static/" + chunk.File
	}

	// fallback if file not found in manifest
	return "/static/" + path
}
