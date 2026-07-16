package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders adds common HTTP security headers to responses.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		// HSTS: only set over HTTPS. Adjust max-age as needed.
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Header("Server", "")

		// Content-Security-Policy for the API (frontend is served separately by Vite).
		// Restrictive defaults: no external resources, no frames, no form actions.
		// Swagger UI needs relaxed CSP to load its own CSS/JS/images.
		if isSwaggerPath(c.Request.URL.Path) {
			c.Header("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; img-src 'self' data:")
		} else {
			c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; form-action 'none'")
		}

		// Prevent MIME type sniffing from affecting caching behavior
		c.Header("Cache-Control", "no-store, max-age=0")
		c.Next()
	}
}

// isSwaggerPath returns true for Swagger UI requests that need a relaxed CSP.
func isSwaggerPath(path string) bool {
	return strings.HasPrefix(path, "/swagger/")
}
