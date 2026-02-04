package validation

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// Test model for validation
type TestUser struct {
	ID        int    `json:"id" validate:"required,min=1"`
	Username  string `json:"username" validate:"required,min_len=3,max_len=20,alphanum"`
	Email     string `json:"email" validate:"required,email"`
	Age       int    `json:"age" validate:"min:18,max:120"`
	Password  string `json:"password" validate:"required,min_len:8"`
	Phone     string `json:"phone"`
	CreatedAt string `json:"created_at" validate:"datetime"`
}

func TestValidator_Validate(t *testing.T) {
	tests := []struct {
		name     string
		user     TestUser
		expected bool
		errors   []string
	}{
		{
			name: "valid user",
			user: TestUser{
				ID:        1,
				Username:  "john123",
				Email:     "john@example.com",
				Age:       25,
				Password:  "password123",
				CreatedAt: "2023-01-01T12:00:00Z",
			},
			expected: true,
		},
		{
			name: "missing required fields",
			user: TestUser{
				Age: 25,
			},
			expected: false,
			errors:   []string{"id", "username", "email", "password"},
		},
		{
			name: "invalid email",
			user: TestUser{
				ID:       1,
				Username: "john123",
				Email:    "invalid-email",
				Age:      25,
				Password: "password123",
			},
			expected: false,
			errors:   []string{"email"},
		},
		{
			name: "age too low",
			user: TestUser{
				ID:       1,
				Username: "john123",
				Email:    "john@example.com",
				Age:      15,
				Password: "password123",
			},
			expected: false,
			errors:   []string{"age"},
		},
		{
			name: "password too short",
			user: TestUser{
				ID:       1,
				Username: "john123",
				Email:    "john@example.com",
				Age:      25,
				Password: "123",
			},
			expected: false,
			errors:   []string{"password"},
		},
		{
			name: "invalid username (special chars)",
			user: TestUser{
				ID:       1,
				Username: "john@123",
				Email:    "john@example.com",
				Age:      25,
				Password: "password123",
			},
			expected: false,
			errors:   []string{"username"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := New()
			isValid := validator.Validate(tt.user)

			if isValid != tt.expected {
				t.Errorf("Expected valid=%v, got valid=%v", tt.expected, isValid)
			}

			if !tt.expected {
				errors := validator.GetErrors()
				for _, expectedError := range tt.errors {
					if _, exists := errors[expectedError]; !exists {
						t.Errorf("Expected error for field '%s'", expectedError)
					}
				}
			}
		})
	}
}

func TestValidator_ValidateField(t *testing.T) {
	validator := New()

	// Test required field
	if validator.ValidateField("name", "", "required") {
		t.Error("Expected empty required field to fail validation")
	}

	if !validator.ValidateField("name", "John", "required") {
		t.Error("Expected non-empty field to pass required validation")
	}

	// Test min length
	if validator.ValidateField("password", "123", "min_len:8") {
		t.Error("Expected short password to fail min_len validation")
	}

	if !validator.ValidateField("password", "password123", "min_len:8") {
		t.Error("Expected long password to pass min_len validation")
	}

	// Test email
	if validator.ValidateField("email", "invalid", "email") {
		t.Error("Expected invalid email to fail validation")
	}

	if !validator.ValidateField("email", "test@example.com", "email") {
		t.Error("Expected valid email to pass validation")
	}

	// Test numeric range
	if validator.ValidateField("age", "15", "min:18") {
		t.Error("Expected age 15 to fail min:18 validation")
	}

	if !validator.ValidateField("age", "25", "min:18,max:30") {
		t.Error("Expected age 25 to pass min:18,max:30 validation")
	}

	if validator.ValidateField("age", "35", "min:18,max:30") {
		t.Error("Expected age 35 to fail max:30 validation")
	}
}

func TestSchemaValidation(t *testing.T) {
	schema := NewSchema().
		Field("name", Required(), MinLen(2), MaxLen(50)).
		Field("email", Required(), Email()).
		Field("age", Min(18), Max(120)).
		Field("category", OneOf("A", "B", "C")).
		Field("price", Min(0), Max(1000)).
		Field("url", URL())

	// Valid data
	validData := map[string]interface{}{
		"name":     "John Doe",
		"email":    "john@example.com",
		"age":      25,
		"category": "A",
		"price":    99.99,
		"url":      "https://example.com",
	}

	isValid, errors := schema.Validate(validData)
	if !isValid {
		t.Errorf("Expected valid data to pass, got errors: %v", errors)
	}

	// Invalid data
	invalidData := map[string]interface{}{
		"name":     "J",         // Too short
		"email":    "invalid",   // Not an email
		"age":      15,          // Too young
		"category": "D",         // Not in list
		"price":    -10,         // Negative
		"url":      "not-a-url", // Invalid URL
	}

	isValid, errors = schema.Validate(invalidData)
	if isValid {
		t.Error("Expected invalid data to fail validation")
	}

	// Check specific errors
	expectedErrors := []string{"name", "email", "age", "category", "price", "url"}
	for _, field := range expectedErrors {
		if _, exists := errors[field]; !exists {
			t.Errorf("Expected error for field '%s'", field)
		}
	}
}

func TestCustomValidation(t *testing.T) {
	schema := NewSchema().
		Field("username", Required(),
			Custom(func(value interface{}) error {
				username, _ := toString(value)
				if username == "admin" {
					return fmt.Errorf("admin username not allowed")
				}
				return nil
			})).
		Field("password", Required(),
			Custom(func(value interface{}) error {
				password, _ := toString(value)
				if len(password) < 8 {
					return fmt.Errorf("password too short")
				}
				return nil
			}))

	// Test with admin username
	data := map[string]interface{}{
		"username": "admin",
		"password": "12345678",
	}

	isValid, errors := schema.Validate(data)
	if isValid {
		t.Error("Expected admin username to fail validation")
	}
	if _, exists := errors["username"]; !exists {
		t.Error("Expected error for username field")
	}

	// Test with weak password
	data = map[string]interface{}{
		"username": "john",
		"password": "123",
	}

	isValid, errors = schema.Validate(data)
	if isValid {
		t.Error("Expected weak password to fail validation")
	}
	if _, exists := errors["password"]; !exists {
		t.Error("Expected error for password field")
	}
}

func TestTimeValidation(t *testing.T) {
	now := time.Now()
	future := now.AddDate(1, 0, 0)
	past := now.AddDate(-1, 0, 0)

	schema := NewSchema().
		Field("start_date", TimeAfter(now)).
		Field("end_date", TimeBefore(now)).
		Field("birth_date", TimeFormat("2006-01-02")).
		Field("event_time", DateTimeOnly())

	// Valid time data
	validData := map[string]interface{}{
		"start_date": future,
		"end_date":   past,
		"birth_date": "1990-01-01",
		"event_time": now.Format(time.RFC3339),
	}

	isValid, errors := schema.Validate(validData)
	if !isValid {
		t.Errorf("Expected valid time data, got errors: %v", errors)
	}

	// Invalid time data
	invalidData := map[string]interface{}{
		"start_date": past,   // Should be after now
		"end_date":   future, // Should be before now
		"birth_date": "invalid-date",
		"event_time": "invalid-datetime",
	}

	isValid, errors = schema.Validate(invalidData)
	if isValid {
		t.Error("Expected invalid time data to fail")
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test IsValidEmail
	if !IsValidEmail("test@example.com") {
		t.Error("Expected valid email to pass")
	}
	if IsValidEmail("invalid") {
		t.Error("Expected invalid email to fail")
	}

	// Test IsValidURL
	if !IsValidURL("https://example.com") {
		t.Error("Expected valid URL to pass")
	}
	if IsValidURL("not-a-url") {
		t.Error("Expected invalid URL to fail")
	}

	// Test IsValidPhone
	if !IsValidPhone("+1 (555) 123-4567") {
		t.Error("Expected valid phone to pass")
	}
	if IsValidPhone("123") {
		t.Error("Expected invalid phone to fail")
	}

	// Test IsValidCreditCard
	if !IsValidCreditCard("4111111111111111") {
		t.Error("Expected valid credit card to pass")
	}
	if IsValidCreditCard("1234567890123456") {
		t.Error("Expected invalid credit card to fail")
	}

	// Test IsValidPassword
	if valid, _ := IsValidPassword("Password123!", 8); !valid {
		t.Error("Expected strong password to pass")
	}
	if valid, _ := IsValidPassword("123", 8); valid {
		t.Error("Expected weak password to fail")
	}

	// Test IsValidDomain
	if !IsValidDomain("example.com") {
		t.Error("Expected valid domain to pass")
	}
	if IsValidDomain("") {
		t.Error("Expected empty domain to fail")
	}

	// Test IsValidHexColor
	if !IsValidHexColor("#ff0000") {
		t.Error("Expected valid hex color to pass")
	}
	if IsValidHexColor("red") {
		t.Error("Expected invalid hex color to fail")
	}

	// Test IsValidISBN
	if !IsValidISBN("978-3-16-148410-0") {
		t.Error("Expected valid ISBN to pass")
	}
	if IsValidISBN("invalid") {
		t.Error("Expected invalid ISBN to fail")
	}
}

func TestSanitization(t *testing.T) {
	// Test SanitizeString
	dirty := "Hello\x00World<script>alert('xss')</script>"
	clean := SanitizeString(dirty)
	if strings.Contains(clean, "\x00") {
		t.Error("Expected null bytes to be removed")
	}
	if strings.Contains(clean, "<script>") {
		t.Error("Expected HTML tags to be removed")
	}

	// Test SanitizeHTML
	html := "<p>Hello <b>World</b></p>"
	text := SanitizeHTML(html)
	if strings.Contains(text, "<") || strings.Contains(text, ">") {
		t.Error("Expected HTML tags to be removed")
	}
	if !strings.Contains(text, "Hello World") {
		t.Error("Expected text content to be preserved")
	}

	// Test SanitizeEmail
	email := "  USER@EXAMPLE.COM  "
	normalized := SanitizeEmail(email)
	if normalized != "user@example.com" {
		t.Errorf("Expected normalized email, got: %s", normalized)
	}

	// Test NormalizePhone
	phone := "+1 (555) 123-4567"
	normalizedPhone := NormalizePhone(phone)
	if normalizedPhone != "+15551234567" {
		t.Errorf("Expected normalized phone, got: %s", normalizedPhone)
	}
}

func TestErrorHandling(t *testing.T) {
	validator := New()

	// Add multiple errors
	validator.AddError("field1", "Error 1")
	validator.AddError("field1", "Error 2")
	validator.AddError("field2", "Error 3")

	errors := validator.GetErrors()
	if len(errors["field1"]) != 2 {
		t.Errorf("Expected 2 errors for field1, got %d", len(errors["field1"]))
	}

	errorList := validator.GetErrorList()
	if len(errorList) != 3 {
		t.Errorf("Expected 3 total errors, got %d", len(errorList))
	}

	// Test Errors type
	errs := Errors{Errors: errorList}
	errStr := errs.Error()
	if !strings.Contains(errStr, "field1: Error 1") {
		t.Error("Expected error string to contain field1 errors")
	}

	// Test clear
	validator.Clear()
	if validator.HasErrors() {
		t.Error("Expected no errors after clear")
	}
}
