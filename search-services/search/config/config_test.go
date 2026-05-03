package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"yadro.com/course/search/config"
)

func TestMustLoad_FromFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.yaml")

	content := `
log_level: DEBUG
search_address: "localhost:9090"
db_address: "postgres://user:pass@localhost/db"
words_address: "localhost:9091"
broker_address: "nats://localhost:4222"
index_ttl: 45s
`
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0o644))

	cfg := config.MustLoad(configPath)

	assert.Equal(t, "DEBUG", cfg.LogLevel)
	assert.Equal(t, "localhost:9090", cfg.Address)
	assert.Equal(t, 45*time.Second, cfg.IndexTTL)
}

func TestMustLoad_FromEnv(t *testing.T) {
	t.Setenv("LOG_LEVEL", "ERROR")
	t.Setenv("SEARCH_ADDRESS", "env-host:8080")
	t.Setenv("DB_ADDRESS", "env-db:5432")
	t.Setenv("WORDS_ADDRESS", "env-words:9000")
	t.Setenv("BROKER_ADDRESS", "env-nats:4222")
	t.Setenv("INDEX_TTL", "60s")

	// Use non-existent file to trigger env fallback
	cfg := config.MustLoad("/nonexistent/config.yaml")

	assert.Equal(t, "ERROR", cfg.LogLevel)
	assert.Equal(t, "env-host:8080", cfg.Address)
	assert.Equal(t, 60*time.Second, cfg.IndexTTL)
}

func TestMustLoad_Defaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "minimal.yaml")

	// Empty config to trigger defaults
	require.NoError(t, os.WriteFile(configPath, []byte("{}"), 0o644))

	cfg := config.MustLoad(configPath)

	assert.Equal(t, "INFO", cfg.LogLevel)
	assert.Equal(t, "localhost:80", cfg.Address)
	assert.Equal(t, 20*time.Second, cfg.IndexTTL)
}
