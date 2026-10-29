package websites

import (
	"testing"
)

func TestNormalizeAndValidateDomain(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		// Canonical domain tests (Requirement B)
		{
			name:        "bare domain",
			input:       "mywebsite.com",
			expected:    "mywebsite.com",
			expectError: false,
		},
		{
			name:        "domain with www",
			input:       "www.mywebsite.com",
			expected:    "mywebsite.com",
			expectError: false,
		},
		{
			name:        "https with www",
			input:       "https://www.mywebsite.com",
			expected:    "mywebsite.com",
			expectError: false,
		},
		{
			name:        "https without www",
			input:       "https://mywebsite.com",
			expected:    "mywebsite.com",
			expectError: false,
		},
		{
			name:        "http with trailing slash",
			input:       "http://mywebsite.com/",
			expected:    "mywebsite.com",
			expectError: false,
		},

		// Page paths identify routes, not separate website keys.
		{
			name:        "www with about path",
			input:       "www.mywebsite.com/about",
			expected:    "mywebsite.com",
			expectError: false,
		},
		{
			name:        "https with about path and trailing slash",
			input:       "https://www.mywebsite.com/about/",
			expected:    "mywebsite.com",
			expectError: false,
		},
		{
			name:        "www with contact path",
			input:       "www.mywebsite.com/contact",
			expected:    "mywebsite.com",
			expectError: false,
		},
		{
			name:        "nested path segments",
			input:       "https://mywebsite.com/products/sale/promo",
			expected:    "mywebsite.com",
			expectError: false,
		},

		// Localhost and testing URLs (Requirement D)
		{
			name:        "plain localhost",
			input:       "localhost",
			expected:    "localhost",
			expectError: false,
		},
		{
			name:        "localhost with port",
			input:       "localhost:3000",
			expected:    "localhost:3000",
			expectError: false,
		},
		{
			name:        "http localhost with port and path",
			input:       "http://localhost:3000/about",
			expected:    "localhost:3000",
			expectError: false,
		},
		{
			name:        "loopback IP with port and path",
			input:       "http://127.0.0.1:8080/contact",
			expected:    "127.0.0.1:8080",
			expectError: false,
		},

		// Query string and fragment removal
		{
			name:        "url with query params and fragment",
			input:       "https://www.mywebsite.com/about?ref=landing#team",
			expected:    "mywebsite.com",
			expectError: false,
		},

		// Error cases
		{
			name:        "empty string",
			input:       "",
			expected:    "",
			expectError: true,
		},
		{
			name:        "whitespace only",
			input:       "   ",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid domain",
			input:       "not a domain",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid port number",
			input:       "localhost:99999",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeAndValidateDomain(tt.input)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error for %q, got nil result %q", tt.input, result)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error for %q: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Fatalf("expected %q, got %q for input %q", tt.expected, result, tt.input)
				}
			}
		})
	}
}
