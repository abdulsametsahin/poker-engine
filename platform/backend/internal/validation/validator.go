package validation

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// Common validation errors
var (
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrInvalidUsername    = errors.New("invalid username format")
	ErrWeakPassword       = errors.New("password too weak")
	ErrInvalidUUID        = errors.New("invalid UUID format")
	ErrInvalidRange       = errors.New("value out of valid range")
	ErrInvalidEnum        = errors.New("invalid enum value")
	ErrInvalidString      = errors.New("invalid string format")
	ErrStringTooLong      = errors.New("string exceeds maximum length")
	ErrStringTooShort     = errors.New("string below minimum length")
	ErrContainsSQLPattern = errors.New("input contains suspicious SQL patterns")
	ErrContainsXSSPattern = errors.New("input contains suspicious XSS patterns")
)

// Regex patterns for validation
var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,20}$`)
	uuidRegex     = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)

	// SQL injection patterns (common attack vectors)
	sqlPatterns = []string{
		"'", "\"", ";", "--", "/*", "*/", "xp_", "sp_",
		"exec", "execute", "select", "insert", "update", "delete",
		"drop", "create", "alter", "union", "script",
	}

	// XSS patterns
	xssPatterns = []string{
		"<script", "</script", "javascript:", "onerror=", "onload=",
		"<iframe", "</iframe", "<object", "</object", "eval(",
	}
)

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}
	if len(email) > 100 {
		return fmt.Errorf("%w: email must be <= 100 characters", ErrStringTooLong)
	}
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}
	return nil
}

// ValidateUsername validates username format
func ValidateUsername(username string) error {
	if username == "" {
		return errors.New("username is required")
	}
	if len(username) < 3 {
		return fmt.Errorf("%w: username must be >= 3 characters", ErrStringTooShort)
	}
	if len(username) > 20 {
		return fmt.Errorf("%w: username must be <= 20 characters", ErrStringTooLong)
	}
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("%w: username can only contain letters, numbers, underscore, and hyphen", ErrInvalidUsername)
	}
	return nil
}

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if password == "" {
		return errors.New("password is required")
	}
	if len(password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters", ErrWeakPassword)
	}
	if len(password) > 128 {
		return fmt.Errorf("%w: password must be <= 128 characters", ErrStringTooLong)
	}

	// Check for at least one uppercase, one lowercase, and one number
	var hasUpper, hasLower, hasNumber bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber {
		return fmt.Errorf("%w: password must contain at least one uppercase letter, one lowercase letter, and one number", ErrWeakPassword)
	}

	return nil
}

// ValidateUUID validates UUID format
func ValidateUUID(uuid string) error {
	if uuid == "" {
		return errors.New("UUID is required")
	}
	if !uuidRegex.MatchString(uuid) {
		return ErrInvalidUUID
	}
	return nil
}

// ValidateIntRange validates integer is within range
func ValidateIntRange(value, min, max int, fieldName string) error {
	if value < min || value > max {
		return fmt.Errorf("%w: %s must be between %d and %d", ErrInvalidRange, fieldName, min, max)
	}
	return nil
}

// ValidatePositiveInt validates integer is positive
func ValidatePositiveInt(value int, fieldName string) error {
	if value <= 0 {
		return fmt.Errorf("%w: %s must be positive", ErrInvalidRange, fieldName)
	}
	return nil
}

// ValidateNonNegativeInt validates integer is non-negative
func ValidateNonNegativeInt(value int, fieldName string) error {
	if value < 0 {
		return fmt.Errorf("%w: %s must be non-negative", ErrInvalidRange, fieldName)
	}
	return nil
}

// ValidateEnum validates value is in allowed list
func ValidateEnum(value string, allowed []string, fieldName string) error {
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return fmt.Errorf("%w: %s must be one of %v", ErrInvalidEnum, fieldName, allowed)
}

// ValidateStringLength validates string length
func ValidateStringLength(value string, minLen, maxLen int, fieldName string) error {
	if len(value) < minLen {
		return fmt.Errorf("%w: %s must be at least %d characters", ErrStringTooShort, fieldName, minLen)
	}
	if len(value) > maxLen {
		return fmt.Errorf("%w: %s must be at most %d characters", ErrStringTooLong, fieldName, maxLen)
	}
	return nil
}

// SanitizeString removes potentially dangerous characters from input
// This is a defense-in-depth measure; parameterized queries are the primary defense
func SanitizeString(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")
	// Trim whitespace
	input = strings.TrimSpace(input)
	return input
}

// CheckSQLInjection checks for common SQL injection patterns
// Note: This is NOT a replacement for parameterized queries!
// This is defense-in-depth to catch obvious attacks
func CheckSQLInjection(input string) error {
	lower := strings.ToLower(input)
	for _, pattern := range sqlPatterns {
		if strings.Contains(lower, strings.ToLower(pattern)) {
			return fmt.Errorf("%w: contains '%s'", ErrContainsSQLPattern, pattern)
		}
	}
	return nil
}

// CheckXSS checks for common XSS patterns
func CheckXSS(input string) error {
	lower := strings.ToLower(input)
	for _, pattern := range xssPatterns {
		if strings.Contains(lower, strings.ToLower(pattern)) {
			return fmt.Errorf("%w: contains '%s'", ErrContainsXSSPattern, pattern)
		}
	}
	return nil
}

// ValidateSafeString validates and sanitizes a general string input
func ValidateSafeString(input string, minLen, maxLen int, fieldName string) (string, error) {
	// Sanitize first
	sanitized := SanitizeString(input)

	// Check length
	if err := ValidateStringLength(sanitized, minLen, maxLen, fieldName); err != nil {
		return "", err
	}

	// Check for SQL injection patterns (defense in depth)
	if err := CheckSQLInjection(sanitized); err != nil {
		return "", fmt.Errorf("%s: %w", fieldName, err)
	}

	// Check for XSS patterns
	if err := CheckXSS(sanitized); err != nil {
		return "", fmt.Errorf("%s: %w", fieldName, err)
	}

	return sanitized, nil
}

// GameAction validators
var ValidGameActions = []string{"fold", "check", "call", "raise", "allin"}

// ValidateGameAction validates poker game action
func ValidateGameAction(action string) error {
	return ValidateEnum(action, ValidGameActions, "action")
}

// ValidateGameActionAmount validates poker game action amount
func ValidateGameActionAmount(action string, amount int) error {
	// Raise must have positive amount
	if action == "raise" && amount <= 0 {
		return fmt.Errorf("raise action requires positive amount")
	}
	// Amount should be reasonable (max 1 billion chips)
	if amount < 0 || amount > 1000000000 {
		return fmt.Errorf("action amount must be between 0 and 1,000,000,000")
	}
	return nil
}

// Table/Tournament validators

// ValidateTableName validates poker table name
func ValidateTableName(name string) error {
	sanitized, err := ValidateSafeString(name, 1, 100, "table name")
	if err != nil {
		return err
	}
	if sanitized != name {
		return errors.New("table name contains invalid characters")
	}
	return nil
}

// ValidateBlinds validates small and big blind values
func ValidateBlinds(smallBlind, bigBlind int) error {
	if err := ValidatePositiveInt(smallBlind, "small blind"); err != nil {
		return err
	}
	if err := ValidatePositiveInt(bigBlind, "big blind"); err != nil {
		return err
	}
	if bigBlind <= smallBlind {
		return errors.New("big blind must be greater than small blind")
	}
	// Reasonable limits (1 to 1 million)
	if smallBlind > 1000000 || bigBlind > 1000000 {
		return errors.New("blinds must be <= 1,000,000")
	}
	return nil
}

// ValidateMaxPlayers validates max players count
func ValidateMaxPlayers(maxPlayers int) error {
	return ValidateIntRange(maxPlayers, 2, 10, "max players")
}

// ValidateBuyIn validates buy-in amount
func ValidateBuyIn(buyIn int) error {
	if err := ValidateNonNegativeInt(buyIn, "buy-in"); err != nil {
		return err
	}
	// Max 100 million chips
	if buyIn > 100000000 {
		return errors.New("buy-in must be <= 100,000,000")
	}
	return nil
}

// ValidateBuyInRange validates min/max buy-in relationship
func ValidateBuyInRange(minBuyIn, maxBuyIn int) error {
	if minBuyIn < 0 {
		return errors.New("min buy-in must be non-negative")
	}
	if maxBuyIn < 0 {
		return errors.New("max buy-in must be non-negative")
	}
	if maxBuyIn > 0 && minBuyIn > maxBuyIn {
		return errors.New("min buy-in must be <= max buy-in")
	}
	return nil
}

// ValidateTournamentName validates tournament name
func ValidateTournamentName(name string) error {
	sanitized, err := ValidateSafeString(name, 1, 100, "tournament name")
	if err != nil {
		return err
	}
	if sanitized != name {
		return errors.New("tournament name contains invalid characters")
	}
	return nil
}

// ValidateTournamentPlayers validates min/max player counts
func ValidateTournamentPlayers(minPlayers, maxPlayers int) error {
	if err := ValidateIntRange(minPlayers, 2, 1000, "min players"); err != nil {
		return err
	}
	if err := ValidateIntRange(maxPlayers, 2, 1000, "max players"); err != nil {
		return err
	}
	if minPlayers > maxPlayers {
		return errors.New("min players must be <= max players")
	}
	return nil
}

// ValidateStartingChips validates tournament starting chips
func ValidateStartingChips(chips int) error {
	return ValidateIntRange(chips, 100, 1000000000, "starting chips")
}
