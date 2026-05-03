package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"yadro.com/course/api/config"
)

func TestMustLoad_FromFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.yaml")
	
	content := `
log_level: DEBUG
search_concurrency: 5
search_rate: 10
api_server:
  address: "localhost:8080"
  timeout: 10s
words_address: "words:81"
update_address: "update:82"
search_address: "search:83"
token_ttl: 48h
`
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0644))
	
	cfg := config.MustLoad(configPath)
	
	assert.Equal(t, "DEBUG", cfg.LogLevel)
	assert.Equal(t, 5, cfg.SearchConcurrency)
	assert.Equal(t, 10, cfg.SearchRate)
	assert.Equal(t, "localhost:8080", cfg.HTTPConfig.Address)
	assert.Equal(t, 10*time.Second, cfg.HTTPConfig.Timeout)
	assert.Equal(t, 48*time.Hour, cfg.TokenTTL)
}

func TestMustLoad_Defaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "minimal.yaml")
	
	require.NoError(t, os.WriteFile(configPath, []byte("{}"), 0644))
	
	cfg := config.MustLoad(configPath)
	
	assert.Equal(t, "DEBUG", cfg.LogLevel)
	assert.Equal(t, "localhost:80", cfg.HTTPConfig.Address)
	assert.Equal(t, 5*time.Second, cfg.HTTPConfig.Timeout)
	assert.Equal(t, 1, cfg.SearchConcurrency)
	assert.Equal(t, 1, cfg.SearchRate)
	assert.Equal(t, 24*time.Hour, cfg.TokenTTL)
}

func TestMustLoad_PartialConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "partial.yaml")
	
	// Only override some fields, rest should use defaults
	content := `
log_level: WARN
api_server:
  address: "0.0.0.0:9000"
`
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0644))
	
	cfg := config.MustLoad(configPath)
	
	assert.Equal(t, "WARN", cfg.LogLevel)
	assert.Equal(t, "0.0.0.0:9000", cfg.HTTPConfig.Address)
	// Defaults should still apply
	assert.Equal(t, 5*time.Second, cfg.HTTPConfig.Timeout)
	assert.Equal(t, 1, cfg.SearchConcurrency)
}
