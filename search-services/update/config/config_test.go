package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"yadro.com/course/update/config"
)

func TestMustLoad_FromFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.yaml")
	
	content := `
log_level: DEBUG
update_address: "localhost:9090"
db_address: "postgres://user:pass@localhost/db"
words_address: "localhost:9091"
broker_address: "nats://localhost:4222"
https://xkcd.com:
  url: "https://xkcd.com"
  concurrency: 5
  timeout: 30s
  check_period: 2h
`
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0644))
	
	cfg := config.MustLoad(configPath)
	
	assert.Equal(t, "DEBUG", cfg.LogLevel)
	assert.Equal(t, "localhost:9090", cfg.Address)
	assert.Equal(t, 5, cfg.XKCD.Concurrency)
	assert.Equal(t, 30*time.Second, cfg.XKCD.Timeout)
}

func TestMustLoad_FromEnv(t *testing.T) {
	t.Setenv("LOG_LEVEL", "WARN")
	t.Setenv("UPDATE_ADDRESS", "env-update:8080")
	t.Setenv("DB_ADDRESS", "env-db:5432")
	t.Setenv("WORDS_ADDRESS", "env-words:9000")
	t.Setenv("BROKER_ADDRESS", "env-nats:4222")
	t.Setenv("XKCD_URL", "https://env.xkcd.com")
	t.Setenv("XKCD_CONCURRENCY", "10")
	t.Setenv("XKCD_TIMEOUT", "15s")
	t.Setenv("XKCD_CHECK_PERIOD", "30m")
	
	cfg := config.MustLoad("/nonexistent/config.yaml")
	
	assert.Equal(t, "WARN", cfg.LogLevel)
	assert.Equal(t, "env-update:8080", cfg.Address)
	assert.Equal(t, 10, cfg.XKCD.Concurrency)
	assert.Equal(t, 15*time.Second, cfg.XKCD.Timeout)
}

func TestMustLoad_Defaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "minimal.yaml")
	
	require.NoError(t, os.WriteFile(configPath, []byte("{}"), 0644))
	
	cfg := config.MustLoad(configPath)
	
	assert.Equal(t, "INFO", cfg.LogLevel)
	assert.Equal(t, "localhost:80", cfg.Address)
	assert.Equal(t, 1, cfg.XKCD.Concurrency)
	assert.Equal(t, 10*time.Second, cfg.XKCD.Timeout)
	assert.Equal(t, "xkcd.com", cfg.XKCD.URL)
}
