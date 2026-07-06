package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ncosentino/google-search-console-mcp/go/internal/config"
)

// TestResolve_FlagTakesPriorityOverEverything confirms the --service-account-file
// flag value wins even when the environment variables are also set.
func TestResolve_FlagTakesPriorityOverEverything(t *testing.T) {
	dir := t.TempDir()
	flagPath := filepath.Join(dir, "flag-creds.json")
	if err := os.WriteFile(flagPath, []byte(`{"source":"flag"}`), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	t.Setenv("GOOGLE_SERVICE_ACCOUNT_FILE", "")
	t.Setenv("GOOGLE_SERVICE_ACCOUNT_JSON", `{"source":"env-json"}`)

	got := config.Resolve(flagPath)
	if string(got.ServiceAccountJSON) != `{"source":"flag"}` {
		t.Errorf("ServiceAccountJSON = %q, want the flag file's content", got.ServiceAccountJSON)
	}
}

// TestResolve_EnvFileTakesPriorityOverEnvJSON confirms GOOGLE_SERVICE_ACCOUNT_FILE
// wins over GOOGLE_SERVICE_ACCOUNT_JSON when no flag is given.
func TestResolve_EnvFileTakesPriorityOverEnvJSON(t *testing.T) {
	dir := t.TempDir()
	envFilePath := filepath.Join(dir, "env-file-creds.json")
	if err := os.WriteFile(envFilePath, []byte(`{"source":"env-file"}`), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	t.Setenv("GOOGLE_SERVICE_ACCOUNT_FILE", envFilePath)
	t.Setenv("GOOGLE_SERVICE_ACCOUNT_JSON", `{"source":"env-json"}`)

	got := config.Resolve("")
	if string(got.ServiceAccountJSON) != `{"source":"env-file"}` {
		t.Errorf("ServiceAccountJSON = %q, want the env-file's content", got.ServiceAccountJSON)
	}
}

// TestResolve_EnvJSONTakesPriorityOverDotEnv confirms GOOGLE_SERVICE_ACCOUNT_JSON
// wins over a .env file when no flag or GOOGLE_SERVICE_ACCOUNT_FILE is given.
func TestResolve_EnvJSONTakesPriorityOverDotEnv(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile(".env", []byte("GOOGLE_SERVICE_ACCOUNT_JSON="+`{"source":"dotenv"}`), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	t.Setenv("GOOGLE_SERVICE_ACCOUNT_FILE", "")
	t.Setenv("GOOGLE_SERVICE_ACCOUNT_JSON", `{"source":"env-json"}`)

	got := config.Resolve("")
	if string(got.ServiceAccountJSON) != `{"source":"env-json"}` {
		t.Errorf("ServiceAccountJSON = %q, want the env var's content", got.ServiceAccountJSON)
	}
}

// TestResolve_FallsBackToDotEnvJSONLine confirms a bare .env file with a
// GOOGLE_SERVICE_ACCOUNT_JSON= line is used when nothing higher-priority is set.
func TestResolve_FallsBackToDotEnvJSONLine(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile(".env", []byte("GOOGLE_SERVICE_ACCOUNT_JSON="+`{"source":"dotenv"}`), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	t.Setenv("GOOGLE_SERVICE_ACCOUNT_FILE", "")
	t.Setenv("GOOGLE_SERVICE_ACCOUNT_JSON", "")

	got := config.Resolve("")
	if string(got.ServiceAccountJSON) != `{"source":"dotenv"}` {
		t.Errorf("ServiceAccountJSON = %q, want the .env file's content", got.ServiceAccountJSON)
	}
}

// TestResolve_FallsBackToDotEnvFileLine confirms a .env file with a
// GOOGLE_SERVICE_ACCOUNT_FILE= line pointing at another file is followed when
// nothing higher-priority is set.
func TestResolve_FallsBackToDotEnvFileLine(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	referencedPath := filepath.Join(dir, "referenced-creds.json")
	if err := os.WriteFile(referencedPath, []byte(`{"source":"dotenv-file"}`), 0o600); err != nil {
		t.Fatalf("WriteFile (referenced): %v", err)
	}
	if err := os.WriteFile(".env", []byte("GOOGLE_SERVICE_ACCOUNT_FILE="+referencedPath), 0o600); err != nil {
		t.Fatalf("WriteFile (.env): %v", err)
	}

	t.Setenv("GOOGLE_SERVICE_ACCOUNT_FILE", "")
	t.Setenv("GOOGLE_SERVICE_ACCOUNT_JSON", "")

	got := config.Resolve("")
	if string(got.ServiceAccountJSON) != `{"source":"dotenv-file"}` {
		t.Errorf("ServiceAccountJSON = %q, want the file referenced by .env", got.ServiceAccountJSON)
	}
}

// TestResolve_NoSourceAvailable_ReturnsEmptyConfig confirms Resolve degrades to an
// empty Config (not a panic or error) when no credential source is available at all --
// main() is responsible for treating this as a fatal startup condition.
func TestResolve_NoSourceAvailable_ReturnsEmptyConfig(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	t.Setenv("GOOGLE_SERVICE_ACCOUNT_FILE", "")
	t.Setenv("GOOGLE_SERVICE_ACCOUNT_JSON", "")

	got := config.Resolve("")
	if len(got.ServiceAccountJSON) != 0 {
		t.Errorf("ServiceAccountJSON = %q, want empty", got.ServiceAccountJSON)
	}
}

// TestResolve_UnreadableFlagFile_FallsThroughToNextSource confirms a
// --service-account-file value pointing at a nonexistent file is treated as
// absent (falls through to the next source) rather than a fatal error --
// matching Resolve's documented behavior of trying each source in turn.
func TestResolve_UnreadableFlagFile_FallsThroughToNextSource(t *testing.T) {
	t.Setenv("GOOGLE_SERVICE_ACCOUNT_FILE", "")
	t.Setenv("GOOGLE_SERVICE_ACCOUNT_JSON", `{"source":"env-json"}`)

	got := config.Resolve(filepath.Join(t.TempDir(), "does-not-exist.json"))
	if string(got.ServiceAccountJSON) != `{"source":"env-json"}` {
		t.Errorf("ServiceAccountJSON = %q, want fallthrough to the env var's content", got.ServiceAccountJSON)
	}
}
