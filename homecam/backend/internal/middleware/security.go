package middleware

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// CORS middleware handles Cross-Origin Resource Sharing.
// Wildcard (*) origins are allowed only without credentials, per the CORS spec.
func CORS(allowedOrigins string) gin.HandlerFunc {
	allowAll := strings.TrimSpace(allowedOrigins) == "*"
	var originList []string
	if !allowAll {
		for _, o := range strings.Split(allowedOrigins, ",") {
			if t := strings.TrimSpace(o); t != "" {
				originList = append(originList, t)
			}
		}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if allowAll {
			// Wildcard must not be combined with credentials — browsers reject it.
			c.Header("Access-Control-Allow-Origin", "*")
		} else {
			matched := false
			for _, o := range originList {
				if o == origin {
					matched = true
					break
				}
			}
			if matched {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
				c.Header("Access-Control-Allow-Credentials", "true")
			}
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// SecureHeaders middleware adds security-related HTTP headers
func SecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Enable XSS filter
		c.Header("X-XSS-Protection", "1; mode=block")

		// Control referrer information
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy
		c.Header("Content-Security-Policy", "default-src 'self'")

		// Strict Transport Security (for HTTPS)
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		c.Next()
	}
}

// Logger middleware provides request logging
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		log.Printf("[%s] %s %s %d %v",
			method,
			path,
			clientIP,
			statusCode,
			latency,
		)
	}
}

// rateLimiter is a simple in-memory sliding-window rate limiter.
// It is per-process; for multi-replica deployments use a Redis-backed limiter.
type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	max      int
	window   time.Duration
}

func newRateLimiter(maxRequests int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		requests: make(map[string][]time.Time),
		max:      maxRequests,
		window:   window,
	}
	go func() {
		ticker := time.NewTicker(window)
		defer ticker.Stop()
		for range ticker.C {
			rl.cleanup()
		}
	}()
	return rl
}

// RateLimiter returns a middleware that limits requests per IP address.
// Each call creates an independent limiter instance (no shared global state).
func RateLimiter(maxRequests int, window time.Duration) gin.HandlerFunc {
	rl := newRateLimiter(maxRequests, window)

	return func(c *gin.Context) {
		if !rl.allow(c.ClientIP()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"detail": "Too many requests. Please try again later.",
			})
			return
		}
		c.Next()
	}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Collect only in-window timestamps, reusing the existing backing array.
	existing := rl.requests[ip]
	valid := existing[:0]
	for _, t := range existing {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.max {
		rl.requests[ip] = valid
		return false
	}

	rl.requests[ip] = append(valid, now)
	return true
}

func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-rl.window)
	for ip, reqs := range rl.requests {
		var active []time.Time
		for _, t := range reqs {
			if t.After(cutoff) {
				active = append(active, t)
			}
		}
		if len(active) == 0 {
			delete(rl.requests, ip)
		} else {
			rl.requests[ip] = active
		}
	}
}
