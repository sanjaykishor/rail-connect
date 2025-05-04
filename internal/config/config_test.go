package config

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

// MockFileReader implements FileReader for testing
type MockFileReader struct {
	files map[string][]byte
}

func (m MockFileReader) ReadFile(filename string) ([]byte, error) {
	if data, ok := m.files[filename]; ok {
		return data, nil
	}
	return nil, errors.New("file not found")
}

func TestLoadConfig(t *testing.T) {
	// Create mock data
	mockReader := MockFileReader{
		files: map[string][]byte{
			"config.yaml": []byte(`
server:
  port: ":50051"
log_level: "info"
sections:
  - name: "A"
    max_seats: 10
  - name: "B"
    max_seats: 20
stations:
  London-France: 20.00`),
		},
	}

	// Test loading a valid configuration file
	cfg, err := LoadConfig("config.yaml", mockReader)
	assert.NoError(t, err, "Should not return an error when loading a valid config file")
	assert.NotNil(t, cfg, "Config should not be nil")
	assert.Equal(t, ":50051", cfg.Server.Port, "Server port should be :50051")
	assert.Equal(t, 2, len(cfg.Sections), "There should be 2 sections in the config")
	assert.Equal(t, "A", cfg.Sections[0].Name, "First section should be A")
	assert.Equal(t, 20, cfg.Sections[1].MaxSeats, "Second section should have 20 max seats")
	assert.Equal(t, 20.00, cfg.Stations["London-France"], "London-France should have a price of 20.00")


	// Test loading an invalid configuration file
	_, err = LoadConfig("invalid_config.yaml", mockReader)
	assert.Error(t, err, "Should return an error when loading an invalid config file")
}

func TestNewLogger(t *testing.T) {
	// Test creating a logger with different log levels
	logger := NewLogger("debug")
	assert.NotNil(t, logger, "Logger should not be nil")

	logger = NewLogger("info")
	assert.NotNil(t, logger, "Logger should not be nil")

	logger = NewLogger("warn")
	assert.NotNil(t, logger, "Logger should not be nil")

	logger = NewLogger("error")
	assert.NotNil(t, logger, "Logger should not be nil")

	logger = NewLogger("invalid")
	assert.NotNil(t, logger, "Logger should not be nil")
}
