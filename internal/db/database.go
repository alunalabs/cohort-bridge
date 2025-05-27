package db

// Database defines the interface for all database types.
type Database interface {
	Get(key string) (string, error)
	List(start, size int) ([]string, error)
}
