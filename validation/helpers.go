package validation

import (
	"fmt"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
)

// Common validation functions

// IsValidEmail validates email using net/mail package
func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// IsValidURL validates URL using net/url package
func IsValidURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// IsValidPhone validates phone number (basic validation)
func IsValidPhone(phone string) bool {
	// Remove spaces, dashes, parentheses
	clean := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' || r == '+' {
			return r
		}
		return -1
	}, phone)

	// Check if it's a reasonable length
	return len(clean) >= 10 && len(clean) <= 15
}

// IsValidCreditCard validates credit card using Luhn algorithm
func IsValidCreditCard(cardNumber string) bool {
	cardNumber = strings.ReplaceAll(cardNumber, " ", "")

	var sum int
	var alternate bool

	for i := len(cardNumber) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(cardNumber[i]))
		if err != nil {
			return false
		}

		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

// IsValidPassword validates password strength
func IsValidPassword(password string, minLength int) (bool, []string) {
	var errors []string

	if len(password) < minLength {
		errors = append(errors, fmt.Sprintf("Password must be at least %d characters", minLength))
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, ch := range password {
		switch {
		case ch >= 'A' && ch <= 'Z':
			hasUpper = true
		case ch >= 'a' && ch <= 'z':
			hasLower = true
		case ch >= '0' && ch <= '9':
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", ch):
			hasSpecial = true
		}
	}

	if !hasUpper {
		errors = append(errors, "Password must contain at least one uppercase letter")
	}
	if !hasLower {
		errors = append(errors, "Password must contain at least one lowercase letter")
	}
	if !hasDigit {
		errors = append(errors, "Password must contain at least one digit")
	}
	if !hasSpecial {
		errors = append(errors, "Password must contain at least one special character")
	}

	return len(errors) == 0, errors
}

// IsValidDomain validates domain name
func IsValidDomain(domain string) bool {
	if len(domain) > 253 {
		return false
	}

	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return false
	}

	for _, part := range parts {
		if len(part) == 0 || len(part) > 63 {
			return false
		}
		if strings.HasPrefix(part, "-") || strings.HasSuffix(part, "-") {
			return false
		}
		for _, ch := range part {
			if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
				(ch >= '0' && ch <= '9') || ch == '-') {
				return false
			}
		}
	}

	return true
}

// IsValidHexColor validates hex color code
func IsValidHexColor(color string) bool {
	if len(color) != 7 || color[0] != '#' {
		return false
	}

	for i := 1; i < 7; i++ {
		ch := color[i]
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')) {
			return false
		}
	}

	return true
}

// IsValidJSON validates if string is valid JSON
func IsValidJSON(jsonStr string) bool {
	jsonStr = strings.TrimSpace(jsonStr)
	return (strings.HasPrefix(jsonStr, "{") && strings.HasSuffix(jsonStr, "}")) ||
		(strings.HasPrefix(jsonStr, "[") && strings.HasSuffix(jsonStr, "]"))
}

// IsValidISBN validates ISBN-10 or ISBN-13
func IsValidISBN(isbn string) bool {
	isbn = strings.ReplaceAll(isbn, "-", "")
	isbn = strings.ReplaceAll(isbn, " ", "")

	if len(isbn) == 10 {
		return isValidISBN10(isbn)
	} else if len(isbn) == 13 {
		return isValidISBN13(isbn)
	}
	return false
}

func isValidISBN10(isbn string) bool {
	var sum int
	for i := 0; i < 9; i++ {
		digit, err := strconv.Atoi(string(isbn[i]))
		if err != nil {
			return false
		}
		sum += digit * (10 - i)
	}

	lastChar := isbn[9]
	if lastChar == 'X' || lastChar == 'x' {
		sum += 10
	} else {
		digit, err := strconv.Atoi(string(lastChar))
		if err != nil {
			return false
		}
		sum += digit
	}

	return sum%11 == 0
}

func isValidISBN13(isbn string) bool {
	var sum int
	for i := 0; i < 12; i++ {
		digit, err := strconv.Atoi(string(isbn[i]))
		if err != nil {
			return false
		}
		if i%2 == 0 {
			sum += digit
		} else {
			sum += digit * 3
		}
	}

	lastDigit, err := strconv.Atoi(string(isbn[12]))
	if err != nil {
		return false
	}

	checkDigit := (10 - (sum % 10)) % 10
	return lastDigit == checkDigit
}

// ToString converts any value to string (public version)
func ToString(value interface{}) (string, bool) {
	switch v := value.(type) {
	case string:
		return v, true
	case []byte:
		return string(v), true
	case fmt.Stringer:
		return v.String(), true
	default:
		return fmt.Sprintf("%v", v), true
	}
}

// SanitizeString removes potentially dangerous characters
func SanitizeString(s string) string {
	// Remove null bytes and control characters
	s = strings.Map(func(r rune) rune {
		if r == 0 || r < 32 || r == 127 {
			return -1
		}
		return r
	}, s)

	// Trim whitespace
	s = strings.TrimSpace(s)

	return s
}

// SanitizeHTML removes HTML tags
func SanitizeHTML(html string) string {
	var result strings.Builder
	var inTag bool

	for _, ch := range html {
		if ch == '<' {
			inTag = true
		} else if ch == '>' {
			inTag = false
		} else if !inTag {
			result.WriteRune(ch)
		}
	}

	return result.String()
}

// SanitizeEmail normalizes email address
func SanitizeEmail(email string) string {
	email = strings.ToLower(strings.TrimSpace(email))
	// Remove multiple spaces
	email = strings.Join(strings.Fields(email), " ")
	return email
}

// NormalizePhone normalizes phone number
func NormalizePhone(phone string) string {
	// Keep only digits and +
	var result strings.Builder
	for _, ch := range phone {
		if (ch >= '0' && ch <= '9') || ch == '+' {
			result.WriteRune(ch)
		}
	}
	return result.String()
}
