package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database struct {
		Type              string   `yaml:"type"`
		Host              string   `yaml:"host"`
		Port              int      `yaml:"port"`
		User              string   `yaml:"user"`
		Password          string   `yaml:"password"`
		DBName            string   `yaml:"dbname"`
		Table             string   `yaml:"table"`
		Filename          string   `yaml:"filename"`
		Fields            []string `yaml:"fields"`
		RandomBitsPercent float64  `yaml:"random_bits_percent"`
		IsTokenized       bool     `yaml:"is_tokenized"`   // Whether the data is already tokenized
		TokenizedFile     string   `yaml:"tokenized_file"` // Path to tokenized data file (JSON/CSV)
	} `yaml:"database"`
	Peer struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"peer"`
	Security struct {
		AllowedIPs      []string `yaml:"allowed_ips"`        // Whitelist of allowed IP addresses
		RequireIPCheck  bool     `yaml:"require_ip_check"`   // Whether to enforce IP whitelist
		MaxConnections  int      `yaml:"max_connections"`    // Maximum concurrent connections
		RateLimitPerMin int      `yaml:"rate_limit_per_min"` // Max connections per minute per IP
	} `yaml:"security"`
	Timeouts struct {
		ConnectionTimeout time.Duration `yaml:"connection_timeout"` // Connection establishment timeout
		ReadTimeout       time.Duration `yaml:"read_timeout"`       // Read operation timeout
		WriteTimeout      time.Duration `yaml:"write_timeout"`      // Write operation timeout
		IdleTimeout       time.Duration `yaml:"idle_timeout"`       // Connection idle timeout
		HandshakeTimeout  time.Duration `yaml:"handshake_timeout"`  // Protocol handshake timeout
	} `yaml:"timeouts"`
	Logging struct {
		Level        string `yaml:"level"`         // Log level: debug, info, warn, error
		File         string `yaml:"file"`          // Log file path (empty for stdout)
		MaxSize      int    `yaml:"max_size"`      // Maximum log file size in MB
		MaxBackups   int    `yaml:"max_backups"`   // Maximum number of old log files
		MaxAge       int    `yaml:"max_age"`       // Maximum age of log files in days
		EnableSyslog bool   `yaml:"enable_syslog"` // Enable syslog output
		EnableAudit  bool   `yaml:"enable_audit"`  // Enable audit logging for security events
		AuditFile    string `yaml:"audit_file"`    // Audit log file path
	} `yaml:"logging"`
	ListenPort int    `yaml:"listen_port"`
	PrivateKey string `yaml:"private_key"`
	PublicKey  string `yaml:"public_key"`
}

// SetDefaults sets reasonable default values for new configuration fields
func (c *Config) SetDefaults() {
	// Security defaults
	if len(c.Security.AllowedIPs) == 0 {
		c.Security.AllowedIPs = []string{"127.0.0.1", "::1"} // localhost only by default
	}
	if c.Security.MaxConnections == 0 {
		c.Security.MaxConnections = 10
	}
	if c.Security.RateLimitPerMin == 0 {
		c.Security.RateLimitPerMin = 5
	}

	// Timeout defaults
	if c.Timeouts.ConnectionTimeout == 0 {
		c.Timeouts.ConnectionTimeout = 30 * time.Second
	}
	if c.Timeouts.ReadTimeout == 0 {
		c.Timeouts.ReadTimeout = 60 * time.Second
	}
	if c.Timeouts.WriteTimeout == 0 {
		c.Timeouts.WriteTimeout = 60 * time.Second
	}
	if c.Timeouts.IdleTimeout == 0 {
		c.Timeouts.IdleTimeout = 300 * time.Second // 5 minutes
	}
	if c.Timeouts.HandshakeTimeout == 0 {
		c.Timeouts.HandshakeTimeout = 30 * time.Second
	}

	// Logging defaults
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.MaxSize == 0 {
		c.Logging.MaxSize = 100 // 100MB
	}
	if c.Logging.MaxBackups == 0 {
		c.Logging.MaxBackups = 3
	}
	if c.Logging.MaxAge == 0 {
		c.Logging.MaxAge = 30 // 30 days
	}
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Apply defaults for any missing configuration
	cfg.SetDefaults()

	return &cfg, nil
}
