package db

import (
	"fmt"

	"github.com/auroradata-ai/cohort-bridge/internal/config"
)

// GetDatabaseFromConfig returns a Database object based on the config info.
func GetDatabaseFromConfig(cfg *config.Config) (Database, error) {
	switch cfg.Database.Type {
	case "csv":
		csvPath := cfg.Database.Filename
		if csvPath == "" {
			csvPath = cfg.Database.Host
		}
		if csvPath == "" {
			csvPath = cfg.Database.Table
		}
		if csvPath == "" {
			return nil, fmt.Errorf("CSV filename not specified in config")
		}
		return NewCSVDatabase(csvPath)
	// case "postgres", "postgresql":
	// return NewPostgresDatabase(cfg.Database)
	// Add other database types here as needed
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
	}
}
