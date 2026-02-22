// Package config resolves Google Search Console service account credentials from multiple sources.
// Priority order: CLI flag (file path) > GOOGLE_SERVICE_ACCOUNT_FILE env var >
// GOOGLE_SERVICE_ACCOUNT_JSON env var > .env file.
package config

import (
	"bufio"
	"log/slog"
	"os"
	"strings"
)

const (
	envVarFile = "GOOGLE_SERVICE_ACCOUNT_FILE"
	envVarJSON = "GOOGLE_SERVICE_ACCOUNT_JSON"
	dotEnvFile = ".env"
)

// Config holds resolved configuration values.
type Config struct {
	// ServiceAccountJSON is the raw service account JSON key content.
	ServiceAccountJSON []byte
}

// Resolve returns a Config with service account credentials from the highest-priority source.
// serviceAccountFile is the value of the --service-account-file CLI flag (may be empty).
func Resolve(serviceAccountFile string) Config {
	if serviceAccountFile != "" {
		if data, err := os.ReadFile(serviceAccountFile); err == nil {
			slog.Debug("service account loaded from CLI flag (file path)")
			return Config{ServiceAccountJSON: data}
		} else {
			slog.Error("failed to read service account file from flag",
				"path", serviceAccountFile, "err", err)
		}
	}

	if v := os.Getenv(envVarFile); v != "" {
		if data, err := os.ReadFile(v); err == nil {
			slog.Debug("service account loaded from GOOGLE_SERVICE_ACCOUNT_FILE env var")
			return Config{ServiceAccountJSON: data}
		} else {
			slog.Error("failed to read service account file from env var",
				"path", v, "err", err)
		}
	}

	if v := os.Getenv(envVarJSON); v != "" {
		slog.Debug("service account loaded from GOOGLE_SERVICE_ACCOUNT_JSON env var")
		return Config{ServiceAccountJSON: []byte(v)}
	}

	if v := loadFromDotEnv(); len(v) > 0 {
		slog.Debug("service account loaded from .env file")
		return Config{ServiceAccountJSON: v}
	}

	return Config{}
}

// loadFromDotEnv reads service account credentials from a .env file in the current directory.
// Checks GOOGLE_SERVICE_ACCOUNT_FILE (file path) first, then GOOGLE_SERVICE_ACCOUNT_JSON.
func loadFromDotEnv() []byte {
	f, err := os.Open(dotEnvFile)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		if after, ok := strings.CutPrefix(line, envVarFile+"="); ok {
			path := strings.Trim(after, `"'`)
			if data, err := os.ReadFile(path); err == nil {
				return data
			}
		}
		if after, ok := strings.CutPrefix(line, envVarJSON+"="); ok {
			return []byte(strings.Trim(after, `"'`))
		}
	}
	return nil
}
