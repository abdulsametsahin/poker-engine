package websocket

import (
	"net/http/httptest"
	"os"
	"testing"
)

func TestGetAllowedOrigins_WithEnv(t *testing.T) {
	// Set environment variable
	os.Setenv("ALLOWED_ORIGINS", "http://example.com,https://app.example.com, http://localhost:3000  ")
	defer os.Unsetenv("ALLOWED_ORIGINS")

	origins := getAllowedOrigins()

	expected := []string{
		"http://example.com",
		"https://app.example.com",
		"http://localhost:3000",
	}

	if len(origins) != len(expected) {
		t.Fatalf("Expected %d origins, got %d", len(expected), len(origins))
	}

	for i, origin := range origins {
		if origin != expected[i] {
			t.Errorf("Expected origin %s, got %s", expected[i], origin)
		}
	}
}

func TestGetAllowedOrigins_WithoutEnv(t *testing.T) {
	// Ensure env var is not set
	os.Unsetenv("ALLOWED_ORIGINS")

	origins := getAllowedOrigins()

	// Should default to localhost
	if len(origins) != 2 {
		t.Fatalf("Expected 2 default origins, got %d", len(origins))
	}

	hasLocalhost := false
	has127 := false
	for _, origin := range origins {
		if origin == "http://localhost:3000" {
			hasLocalhost = true
		}
		if origin == "http://127.0.0.1:3000" {
			has127 = true
		}
	}

	if !hasLocalhost {
		t.Error("Expected default origins to include http://localhost:3000")
	}
	if !has127 {
		t.Error("Expected default origins to include http://127.0.0.1:3000")
	}
}

func TestCheckOrigin_AllowedOrigin(t *testing.T) {
	// Set allowed origins for test
	AllowedOrigins = []string{"http://localhost:3000", "https://app.example.com"}

	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	if !checkOrigin(req) {
		t.Error("Expected to allow connection from http://localhost:3000")
	}
}

func TestCheckOrigin_DisallowedOrigin(t *testing.T) {
	// Set allowed origins for test
	AllowedOrigins = []string{"http://localhost:3000", "https://app.example.com"}

	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Origin", "http://evil.com")

	if checkOrigin(req) {
		t.Error("Expected to reject connection from http://evil.com")
	}
}

func TestCheckOrigin_MissingOriginHeader(t *testing.T) {
	// Set allowed origins for test
	AllowedOrigins = []string{"http://localhost:3000"}

	req := httptest.NewRequest("GET", "/ws", nil)
	// No Origin header set

	if checkOrigin(req) {
		t.Error("Expected to reject connection without Origin header")
	}
}

func TestCheckOrigin_EmptyOriginHeader(t *testing.T) {
	// Set allowed origins for test
	AllowedOrigins = []string{"http://localhost:3000"}

	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Origin", "")

	if checkOrigin(req) {
		t.Error("Expected to reject connection with empty Origin header")
	}
}

func TestCheckOrigin_CaseSensitive(t *testing.T) {
	// Set allowed origins for test
	AllowedOrigins = []string{"http://localhost:3000"}

	// Origin header is case-sensitive for the domain
	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Origin", "http://LOCALHOST:3000")

	// This should be rejected because Origin matching is case-sensitive
	if checkOrigin(req) {
		t.Error("Expected origin matching to be case-sensitive")
	}
}

func TestCheckOrigin_ProtocolMismatch(t *testing.T) {
	// Set allowed origins for test
	AllowedOrigins = []string{"https://app.example.com"}

	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Origin", "http://app.example.com") // http instead of https

	if checkOrigin(req) {
		t.Error("Expected to reject connection with different protocol")
	}
}

func TestCheckOrigin_PortMismatch(t *testing.T) {
	// Set allowed origins for test
	AllowedOrigins = []string{"http://localhost:3000"}

	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Origin", "http://localhost:8080") // different port

	if checkOrigin(req) {
		t.Error("Expected to reject connection with different port")
	}
}

func TestCheckOrigin_SubdomainNotAllowed(t *testing.T) {
	// Set allowed origins for test
	AllowedOrigins = []string{"http://example.com"}

	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Origin", "http://sub.example.com") // subdomain

	if checkOrigin(req) {
		t.Error("Expected to reject subdomain when only parent domain is allowed")
	}
}

func TestCheckOrigin_MultipleAllowedOrigins(t *testing.T) {
	// Set multiple allowed origins
	AllowedOrigins = []string{
		"http://localhost:3000",
		"https://app.example.com",
		"https://staging.example.com",
	}

	tests := []struct {
		name     string
		origin   string
		expected bool
	}{
		{"First allowed", "http://localhost:3000", true},
		{"Second allowed", "https://app.example.com", true},
		{"Third allowed", "https://staging.example.com", true},
		{"Not allowed", "https://evil.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/ws", nil)
			req.Header.Set("Origin", tt.origin)

			result := checkOrigin(req)
			if result != tt.expected {
				t.Errorf("For origin %s: expected %v, got %v", tt.origin, tt.expected, result)
			}
		})
	}
}
