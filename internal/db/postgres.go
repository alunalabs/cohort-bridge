package db

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"

	_ "github.com/lib/pq" // PostgreSQL driver
)

type PostgresDatabase struct {
	db        *sql.DB
	tableName string
	columns   []string
	keyColumn string
	mu        sync.RWMutex
}

// NewPostgresDatabase creates a new PostgreSQL database connection.
func NewPostgresDatabase(dbConfig interface{}) (*PostgresDatabase, error) {
	// Type assertion to access fields
	type DatabaseConfig struct {
		Type              string
		Host              string
		Port              int
		User              string
		Password          string
		DBName            string
		Table             string
		Filename          string
		Fields            []string
		RandomBitsPercent float64
		IsTokenized       bool
		TokenizedFile     string
	}

	// Convert to our expected type
	var cfg DatabaseConfig
	switch v := dbConfig.(type) {
	case DatabaseConfig:
		cfg = v
	default:
		// Try to access via reflection-like field access
		// This is a bit hacky but works for our use case
		if reflect.TypeOf(dbConfig).Kind() == reflect.Struct {
			val := reflect.ValueOf(dbConfig)
			cfg.Type = val.FieldByName("Type").String()
			cfg.Host = val.FieldByName("Host").String()
			cfg.Port = int(val.FieldByName("Port").Int())
			cfg.User = val.FieldByName("User").String()
			cfg.Password = val.FieldByName("Password").String()
			cfg.DBName = val.FieldByName("DBName").String()
			cfg.Table = val.FieldByName("Table").String()
		} else {
			return nil, fmt.Errorf("unsupported config type: %T", dbConfig)
		}
	}
	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	pgDB := &PostgresDatabase{
		db:        db,
		tableName: cfg.Table,
	}

	// Get table schema to determine columns and key column
	if err := pgDB.loadTableSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to load table schema: %w", err)
	}

	return pgDB, nil
}

// loadTableSchema retrieves the column information from the table
func (db *PostgresDatabase) loadTableSchema() error {
	query := `
		SELECT column_name 
		FROM information_schema.columns 
		WHERE table_name = $1 
		ORDER BY ordinal_position`

	rows, err := db.db.Query(query, db.tableName)
	if err != nil {
		return fmt.Errorf("failed to query table schema: %w", err)
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var column string
		if err := rows.Scan(&column); err != nil {
			return fmt.Errorf("failed to scan column name: %w", err)
		}
		columns = append(columns, column)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating over columns: %w", err)
	}

	if len(columns) == 0 {
		return fmt.Errorf("table %s has no columns or does not exist", db.tableName)
	}

	db.columns = columns
	db.keyColumn = columns[0] // Use the first column as the key column (similar to CSV)

	return nil
}

// Get returns the row as a map[columnName]value for the given key.
func (db *PostgresDatabase) Get(key string) (map[string]string, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	// Build the SELECT query
	columnList := strings.Join(db.columns, ", ")
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", columnList, db.tableName, db.keyColumn)

	row := db.db.QueryRow(query, key)

	// Create slices to hold the values
	values := make([]interface{}, len(db.columns))
	valuePtrs := make([]interface{}, len(db.columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Scan the row
	if err := row.Scan(valuePtrs...); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("key not found")
		}
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	// Convert to map[string]string
	result := make(map[string]string)
	for i, column := range db.columns {
		if values[i] != nil {
			result[column] = fmt.Sprintf("%v", values[i])
		} else {
			result[column] = ""
		}
	}

	return result, nil
}

// List returns a slice of row maps starting from `start` index, up to `size` entries.
func (db *PostgresDatabase) List(start, size int) ([]map[string]string, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if start < 0 {
		return nil, fmt.Errorf("start index must be non-negative")
	}

	// Build the SELECT query with LIMIT and OFFSET
	columnList := strings.Join(db.columns, ", ")
	query := fmt.Sprintf("SELECT %s FROM %s ORDER BY %s LIMIT $1 OFFSET $2",
		columnList, db.tableName, db.keyColumn)

	rows, err := db.db.Query(query, size, start)
	if err != nil {
		return nil, fmt.Errorf("failed to query rows: %w", err)
	}
	defer rows.Close()

	var result []map[string]string

	for rows.Next() {
		// Create slices to hold the values
		values := make([]interface{}, len(db.columns))
		valuePtrs := make([]interface{}, len(db.columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan the row
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Convert to map[string]string
		row := make(map[string]string)
		for i, column := range db.columns {
			if values[i] != nil {
				row[column] = fmt.Sprintf("%v", values[i])
			} else {
				row[column] = ""
			}
		}

		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return result, nil
}

// Close closes the database connection
func (db *PostgresDatabase) Close() error {
	return db.db.Close()
}
