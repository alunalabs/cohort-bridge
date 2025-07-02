// Package server provides security middleware and connection management
package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/auroradata-ai/cohort-bridge/internal/config"
)

// SecurityManager handles security policies and connection management
type SecurityManager struct {
	config       *config.Config
	currentConns int
	rateLimitMap map[string]*rateLimitInfo
	mutex        sync.RWMutex
}

type rateLimitInfo struct {
	count     int
	resetTime time.Time
}

// NewSecurityManager creates a new security manager
func NewSecurityManager(cfg *config.Config) *SecurityManager {
	return &SecurityManager{
		config:       cfg,
		rateLimitMap: make(map[string]*rateLimitInfo),
	}
}

// ValidateConnection checks if a connection should be allowed
func (sm *SecurityManager) ValidateConnection(remoteAddr string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return fmt.Errorf("invalid remote address: %w", err)
	}

	// Check rate limiting
	if err := sm.checkRateLimit(host); err != nil {
		return err
	}

	return nil
}

// checkRateLimit enforces per-IP rate limiting
func (sm *SecurityManager) checkRateLimit(host string) error {
	now := time.Now()

	info, exists := sm.rateLimitMap[host]
	if !exists || now.After(info.resetTime) {
		// Reset or create new rate limit info
		sm.rateLimitMap[host] = &rateLimitInfo{
			count:     1,
			resetTime: now.Add(time.Minute),
		}
		return nil
	}

	if info.count >= sm.config.Security.RateLimitPerMin {
		return fmt.Errorf("rate limit exceeded for IP %s", host)
	}

	info.count++
	return nil
}

// TrackConnection increments the connection counter
func (sm *SecurityManager) TrackConnection() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.currentConns++
}

// ReleaseConnection decrements the connection counter
func (sm *SecurityManager) ReleaseConnection() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	if sm.currentConns > 0 {
		sm.currentConns--
	}
}

// SecurityMiddleware provides HTTP security middleware
func (sm *SecurityManager) SecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate the connection
		if err := sm.ValidateConnection(r.RemoteAddr); err != nil {
			http.Error(w, "Connection not allowed: "+err.Error(), http.StatusForbidden)
			return
		}

		// Track the connection
		sm.TrackConnection()
		defer sm.ReleaseConnection()

		// Set security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		next.ServeHTTP(w, r)
	})
}

// SecurityContextKey is used for context values
type SecurityContextKey string

const (
	// RemoteIPKey is the context key for remote IP
	RemoteIPKey SecurityContextKey = "remote_ip"
)

// WithSecurityContext adds security information to the request context
func (sm *SecurityManager) WithSecurityContext(r *http.Request) *http.Request {
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	ctx := context.WithValue(r.Context(), RemoteIPKey, host)
	return r.WithContext(ctx)
}

// GetSecurityStats returns current security statistics
func (sm *SecurityManager) GetSecurityStats() map[string]interface{} {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	stats := map[string]interface{}{
		"current_connections": sm.currentConns,
		"rate_limit_per_min":  sm.config.Security.RateLimitPerMin,
		"monitored_ips":       len(sm.rateLimitMap),
	}

	return stats
}
