package validation

import (
	"strings"
	"testing"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"Valid email", "user@example.com", false},
		{"Valid email with subdomain", "user@mail.example.com", false},
		{"Valid email with plus", "user+tag@example.com", false},
		{"Empty email", "", true},
		{"No @", "userexample.com", true},
		{"No domain", "user@", true},
		{"No TLD", "user@example", true},
		{"Too long", strings.Repeat("a", 100) + "@example.com", true},
		{"Invalid characters", "user<script>@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"Valid username", "user123", false},
		{"Valid with underscore", "user_name", false},
		{"Valid with hyphen", "user-name", false},
		{"Minimum length", "abc", false},
		{"Maximum length", "a12345678901234567890", true},  // 21 chars
		{"Too short", "ab", true},
		{"Empty", "", true},
		{"With spaces", "user name", true},
		{"With special chars", "user@name", true},
		{"With unicode", "usér", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUsername() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"Valid strong password", "Password123", false},
		{"Valid with special chars", "Pass@word123", false},
		{"Too short", "Pass1", true},
		{"No uppercase", "password123", true},
		{"No lowercase", "PASSWORD123", true},
		{"No number", "PasswordABC", true},
		{"Empty", "", true},
		{"Too long", strings.Repeat("A", 129) + "a1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		name    string
		uuid    string
		wantErr bool
	}{
		{"Valid UUID", "550e8400-e29b-41d4-a716-446655440000", false},
		{"Valid UUID uppercase", "550E8400-E29B-41D4-A716-446655440000", false},
		{"Invalid format", "not-a-uuid", true},
		{"Missing hyphens", "550e8400e29b41d4a716446655440000", true},
		{"Too short", "550e8400-e29b-41d4-a716", true},
		{"Empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUUID(tt.uuid)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUUID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateIntRange(t *testing.T) {
	tests := []struct {
		name      string
		value     int
		min       int
		max       int
		fieldName string
		wantErr   bool
	}{
		{"Within range", 5, 1, 10, "test", false},
		{"At minimum", 1, 1, 10, "test", false},
		{"At maximum", 10, 1, 10, "test", false},
		{"Below minimum", 0, 1, 10, "test", true},
		{"Above maximum", 11, 1, 10, "test", true},
		{"Negative in positive range", -5, 0, 10, "test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIntRange(tt.value, tt.min, tt.max, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIntRange() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateGameAction(t *testing.T) {
	tests := []struct {
		name    string
		action  string
		wantErr bool
	}{
		{"Valid fold", "fold", false},
		{"Valid check", "check", false},
		{"Valid call", "call", false},
		{"Valid raise", "raise", false},
		{"Valid allin", "allin", false},
		{"Invalid action", "invalid", true},
		{"Empty", "", true},
		{"Case sensitive", "Fold", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGameAction(tt.action)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGameAction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateGameActionAmount(t *testing.T) {
	tests := []struct {
		name    string
		action  string
		amount  int
		wantErr bool
	}{
		{"Raise with positive amount", "raise", 100, false},
		{"Raise with zero amount", "raise", 0, true},
		{"Raise with negative amount", "raise", -100, true},
		{"Call with zero amount", "call", 0, false},
		{"Fold with zero amount", "fold", 0, false},
		{"Exceeds maximum", "raise", 2000000000, true},
		{"Negative amount", "call", -50, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGameActionAmount(tt.action, tt.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGameActionAmount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckSQLInjection(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Clean input", "hello world", false},
		{"Single quote", "it's fine", true},
		{"Double quote", "he said \"hello\"", true},
		{"SQL comment", "text -- comment", true},
		{"SQL keyword SELECT", "SELECT * FROM users", true},
		{"SQL keyword DROP", "DROP TABLE users", true},
		{"SQL UNION", "UNION SELECT password", true},
		{"Clean with numbers", "user123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckSQLInjection(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckSQLInjection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckXSS(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Clean input", "hello world", false},
		{"Script tag", "<script>alert('xss')</script>", true},
		{"JavaScript protocol", "javascript:alert(1)", true},
		{"Onerror handler", "<img onerror='alert(1)'>", true},
		{"Iframe tag", "<iframe src='evil.com'>", true},
		{"Clean HTML-like", "less than < and greater than >", false},
		{"Clean with brackets", "array[0]", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckXSS(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckXSS() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateBlinds(t *testing.T) {
	tests := []struct {
		name       string
		smallBlind int
		bigBlind   int
		wantErr    bool
	}{
		{"Valid 1/2", 1, 2, false},
		{"Valid 5/10", 5, 10, false},
		{"Valid 50/100", 50, 100, false},
		{"Big blind equals small blind", 10, 10, true},
		{"Big blind less than small blind", 10, 5, true},
		{"Zero small blind", 0, 10, true},
		{"Zero big blind", 10, 0, true},
		{"Negative blinds", -5, 10, true},
		{"Exceeds maximum", 2000000, 3000000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBlinds(tt.smallBlind, tt.bigBlind)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBlinds() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateMaxPlayers(t *testing.T) {
	tests := []struct {
		name       string
		maxPlayers int
		wantErr    bool
	}{
		{"Valid 2 players", 2, false},
		{"Valid 6 players", 6, false},
		{"Valid 10 players", 10, false},
		{"Too few", 1, true},
		{"Too many", 11, true},
		{"Zero", 0, true},
		{"Negative", -5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMaxPlayers(tt.maxPlayers)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMaxPlayers() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTournamentPlayers(t *testing.T) {
	tests := []struct {
		name       string
		minPlayers int
		maxPlayers int
		wantErr    bool
	}{
		{"Valid 2-100", 2, 100, false},
		{"Valid 10-1000", 10, 1000, false},
		{"Min greater than max", 100, 50, true},
		{"Min too small", 1, 100, true},
		{"Max too large", 2, 1001, true},
		{"Both invalid", 1, 1001, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTournamentPlayers(tt.minPlayers, tt.maxPlayers)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTournamentPlayers() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Clean string", "hello", "hello"},
		{"With whitespace", "  hello  ", "hello"},
		{"With null byte", "hello\x00world", "helloworld"},
		{"Multiple spaces", "hello    world", "hello    world"}, // Only trims edges
		{"Empty", "", ""},
		{"Only whitespace", "   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestValidateSafeString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		minLen    int
		maxLen    int
		fieldName string
		wantErr   bool
	}{
		{"Valid string", "hello", 1, 10, "test", false},
		{"With whitespace", "  hello  ", 1, 10, "test", false},
		{"Too short after sanitize", "   ", 1, 10, "test", true},
		{"Too long", "hello world long string", 1, 10, "test", true},
		{"With SQL injection", "'; DROP TABLE users; --", 1, 100, "test", true},
		{"With XSS", "<script>alert(1)</script>", 1, 100, "test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateSafeString(tt.input, tt.minLen, tt.maxLen, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSafeString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
