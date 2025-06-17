package server

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/auroradata-ai/cohort-bridge/internal/config"
)

// SecurityManager handles IP whitelisting and rate limiting
type SecurityManager struct {
	config       *config.Config
	allowedIPs   map[string]bool
	connections  map[string]int
	lastAccess   map[string]time.Time
	rateLimiter  map[string][]time.Time
	mu           sync.RWMutex
	currentConns int
}

// NewSecurityManager creates a new security manager
func NewSecurityManager(cfg *config.Config) *SecurityManager {
	sm := &SecurityManager{
		config:      cfg,
		allowedIPs:  make(map[string]bool),
		connections: make(map[string]int),
		lastAccess:  make(map[string]time.Time),
		rateLimiter: make(map[string][]time.Time),
	}

	// Initialize allowed IPs
	for _, ip := range cfg.Security.AllowedIPs {
		sm.allowedIPs[ip] = true
	}

	// Start cleanup routine
	go sm.cleanupRoutine()

	return sm
}

// IsIPAllowed checks if an IP address is allowed to connect
func (sm *SecurityManager) IsIPAllowed(remoteAddr string) bool {
	if !sm.config.Security.RequireIPCheck {
		return true // IP checking disabled
	}

	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		Debug("Failed to parse remote address %s: %v", remoteAddr, err)
		return false
	}

	sm.mu.RLock()
	allowed := sm.allowedIPs[host]
	sm.mu.RUnlock()

	if !allowed {
		Audit("IP_ACCESS_DENIED", map[string]interface{}{
			"remote_ip": host,
			"reason":    "not_in_whitelist",
		})
	}

	return allowed
}

// CheckRateLimit checks if the IP is within rate limits
func (sm *SecurityManager) CheckRateLimit(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return false
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-time.Minute)

	// Initialize or get existing access times for this IP
	accesses, exists := sm.rateLimiter[host]
	if !exists {
		accesses = make([]time.Time, 0)
	}

	// Remove old entries (older than 1 minute)
	var recentAccesses []time.Time
	for _, accessTime := range accesses {
		if accessTime.After(cutoff) {
			recentAccesses = append(recentAccesses, accessTime)
		}
	}

	// Check if within rate limit
	if len(recentAccesses) >= sm.config.Security.RateLimitPerMin {
		Audit("RATE_LIMIT_EXCEEDED", map[string]interface{}{
			"remote_ip":    host,
			"access_count": len(recentAccesses),
			"limit":        sm.config.Security.RateLimitPerMin,
			"window":       "1_minute",
		})
		return false
	}

	// Add current access
	recentAccesses = append(recentAccesses, now)
	sm.rateLimiter[host] = recentAccesses
	sm.lastAccess[host] = now

	return true
}

// CanAcceptConnection checks if a new connection can be accepted
func (sm *SecurityManager) CanAcceptConnection(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return false
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check maximum connections limit
	if sm.currentConns >= sm.config.Security.MaxConnections {
		Audit("MAX_CONNECTIONS_EXCEEDED", map[string]interface{}{
			"remote_ip":           host,
			"current_connections": sm.currentConns,
			"max_connections":     sm.config.Security.MaxConnections,
		})
		return false
	}

	return true
}

// RecordConnection records a new connection from an IP
func (sm *SecurityManager) RecordConnection(remoteAddr string) {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.connections[host]++
	sm.currentConns++
	sm.lastAccess[host] = time.Now()

	Audit("CONNECTION_ESTABLISHED", map[string]interface{}{
		"remote_ip":           host,
		"connections_from_ip": sm.connections[host],
		"total_connections":   sm.currentConns,
	})
}

// RecordDisconnection records a connection closure
func (sm *SecurityManager) RecordDisconnection(remoteAddr string) {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.connections[host] > 0 {
		sm.connections[host]--
	}
	if sm.currentConns > 0 {
		sm.currentConns--
	}

	Audit("CONNECTION_CLOSED", map[string]interface{}{
		"remote_ip":           host,
		"connections_from_ip": sm.connections[host],
		"total_connections":   sm.currentConns,
	})
}

// ValidateConnection performs all security checks for a new connection
func (sm *SecurityManager) ValidateConnection(remoteAddr string) error {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return fmt.Errorf("invalid remote address: %w", err)
	}

	if !sm.IsIPAllowed(remoteAddr) {
		return fmt.Errorf("IP %s not allowed", host)
	}

	if !sm.CheckRateLimit(remoteAddr) {
		return fmt.Errorf("rate limit exceeded for IP %s", host)
	}

	if !sm.CanAcceptConnection(remoteAddr) {
		return fmt.Errorf("maximum connections exceeded")
	}

	return nil
}

// cleanupRoutine periodically cleans up old entries
func (sm *SecurityManager) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sm.cleanup()
	}
}

// cleanup removes old entries from rate limiter and connection tracking
func (sm *SecurityManager) cleanup() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-time.Hour) // Keep data for 1 hour

	// Clean up rate limiter entries
	for ip, accesses := range sm.rateLimiter {
		var recentAccesses []time.Time
		for _, accessTime := range accesses {
			if accessTime.After(cutoff) {
				recentAccesses = append(recentAccesses, accessTime)
			}
		}

		if len(recentAccesses) == 0 {
			delete(sm.rateLimiter, ip)
			delete(sm.lastAccess, ip)
		} else {
			sm.rateLimiter[ip] = recentAccesses
		}
	}

	Debug("Security cleanup completed: tracking %d IPs", len(sm.rateLimiter))
}

// GetConnectionStats returns current connection statistics
func (sm *SecurityManager) GetConnectionStats() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return map[string]interface{}{
		"total_connections":  sm.currentConns,
		"max_connections":    sm.config.Security.MaxConnections,
		"tracked_ips":        len(sm.rateLimiter),
		"rate_limit_per_min": sm.config.Security.RateLimitPerMin,
		"ip_check_enabled":   sm.config.Security.RequireIPCheck,
		"allowed_ips_count":  len(sm.allowedIPs),
	}
}
