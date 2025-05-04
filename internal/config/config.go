package config

import (
	"fmt"
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Server   ServerConfig       `yaml:"server"`
	LogLevel string             `yaml:"log_level"`
	Sections []SectionConfig    `yaml:"sections"`
	Stations map[string]float64 `yaml:"stations"`
}

// ServerConfig holds the server-specific configuration.
type ServerConfig struct {
	Port string `yaml:"port"`
}

// SectionConfig holds the configuration for each section.
type SectionConfig struct {
	Name     string `yaml:"name"`
	MaxSeats int    `yaml:"max_seats"`
}

// FileReader is an interface for reading files
type FileReader interface {
	ReadFile(filename string) ([]byte, error)
}

type OSFileReader struct{}

func (r OSFileReader) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// LoadConfig loads configuration from a file using the provided FileReader
func LoadConfig(filename string, reader FileReader) (*Config, error) {
	data, err := reader.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// NewLogger initializes a new Zap logger.
func NewLogger(logLevel string) *zap.Logger {
	var level zap.AtomicLevel
	switch logLevel {
	case "debug":
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		level = zap.NewAtomicLevelAt(zap.InfoLevel) // Default to info level
	}

	cfg := zap.Config{
		Encoding:         "json",
		Level:            level,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:   "message",
			LevelKey:     "level",
			TimeKey:      "time",
			CallerKey:    "caller",
			EncodeLevel:  zapcore.LowercaseLevelEncoder,
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}
	logger, err := cfg.Build()
	if err != nil {
		log.Fatalf("failed to initialize zap logger: %v", err)
	}
	return logger
}
