package db

// Database defines the interface for all database types.
type Database interface {
	Get(key string) (map[string]string, error)
	List(start, size int) ([]map[string]string, error)
}
