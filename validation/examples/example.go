package examples

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/selanim/sego/validation"
)

// User represents a user model with validation tags
type User struct {
	ID        int    `json:"id" validate:"required,min=1"`
	Name      string `json:"name" validate:"required,min_len=2,max_len=100"`
	Email     string `json:"email" validate:"required,email"`
	Age       int    `json:"age" validate:"min=18,max=120"`
	Password  string `json:"password" validate:"required,min_len=8"`
	Phone     string `json:"phone" validate:"required"`
	CreatedAt string `json:"created_at" validate:"datetime"`
}

// Product represents a product model
type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
}

func ExampleTagValidation() {
	user := User{
		Name:      "J",
		Email:     "invalid-email",
		Age:       15,
		Password:  "123",
		CreatedAt: "invalid-date",
	}

	validator := validation.New()
	isValid := validator.Validate(user)

	if !isValid {
		errors := validator.GetErrors()
		for field, fieldErrors := range errors {
			fmt.Printf("%s:\n", field)
			for _, err := range fieldErrors {
				fmt.Printf("  - %s\n", err)
			}
		}
	}
}

func ExampleSchemaValidation() {
	// Create a validation schema
	schema := validation.NewSchema().
		Field("name", validation.Required(), validation.MinLen(2), validation.MaxLen(100)).
		Field("email", validation.Required(), validation.Email()).
		Field("age", validation.Min(18), validation.Max(120)).
		Field("password", validation.Required(), validation.MinLen(8)).
		Field("category", validation.OneOf("electronics", "clothing", "books", "home")).
		Field("price", validation.Min(0), validation.Max(10000)).
		Field("expiry_date", validation.TimeAfter(time.Now())).
		Field("website", validation.URL()).
		Field("sku", validation.AlphaNum()).
		Field("ip_address", validation.IPv4()).
		Field("uuid", validation.UUID())

	// Test data
	data := map[string]interface{}{
		"name":        "John",
		"email":       "john@example.com",
		"age":         25,
		"password":    "secure123!",
		"category":    "electronics",
		"price":       99.99,
		"expiry_date": time.Now().AddDate(1, 0, 0),
		"website":     "https://example.com",
		"sku":         "ABC123",
		"ip_address":  "192.168.1.1",
		"uuid":        "123e4567-e89b-12d3-a456-426614174000",
	}

	// Validate
	isValid, errors := schema.Validate(data)
	if isValid {
		fmt.Println("Data is valid!")
	} else {
		fmt.Println("Validation errors:")
		for field, fieldErrors := range errors {
			fmt.Printf("%s: %v\n", field, fieldErrors)
		}
	}
}

func ExampleCustomValidation() {
	// Create schema with custom validation
	schema := validation.NewSchema().
		Field("username", validation.Required(), validation.MinLen(3), validation.MaxLen(20),
			validation.Custom(func(value interface{}) error {
				username, _ := validation.ToString(value)
				if username == "admin" || username == "root" {
					return fmt.Errorf("username '%s' is reserved", username)
				}
				return nil
			}, "Username is reserved")).
		Field("password", validation.Required(),
			validation.Custom(func(value interface{}) error {
				password, _ := validation.ToString(value)
				if valid, errors := validation.IsValidPassword(password, 8); !valid {
					return fmt.Errorf("%s", strings.Join(errors, ", "))
				}
				return nil
			}, "Password does not meet requirements")).
		Field("email", validation.Required(), validation.Email(),
			validation.Custom(func(value interface{}) error {
				email, _ := validation.ToString(value)
				if !strings.HasSuffix(email, "@company.com") {
					return fmt.Errorf("must be a company email")
				}
				return nil
			}))

	data := map[string]interface{}{
		"username": "admin",
		"password": "123",
		"email":    "user@gmail.com",
	}

	isValid, errors := schema.Validate(data)
	fmt.Printf("Is valid: %v\n", isValid)
	fmt.Printf("Errors: %v\n", errors)
}

func ExampleStructValidation() {
	// Product validation schema
	productSchema := validation.NewSchema().
		Field("Name", validation.Required(), validation.MinLen(2), validation.MaxLen(200)).
		Field("Price", validation.Required(),
			validation.Custom(func(value interface{}) error {
				// Convert value to float64
				var price float64
				switch v := value.(type) {
				case float64:
					price = v
				case float32:
					price = float64(v)
				case int:
					price = float64(v)
				case string:
					if f, err := strconv.ParseFloat(v, 64); err == nil {
						price = f
					} else {
						return fmt.Errorf("invalid price format")
					}
				default:
					return fmt.Errorf("price must be a number")
				}

				if price < 0.01 {
					return fmt.Errorf("price must be at least 0.01")
				}
				if price > 1000000 {
					return fmt.Errorf("price cannot exceed 1,000,000")
				}
				return nil
			}, "Price must be between 0.01 and 1,000,000")).
		Field("Category", validation.Required(), validation.OneOf("electronics", "clothing", "books")).
		Field("Description", validation.MaxLen(1000))

	product := Product{
		Name:        "Laptop",
		Price:       0,
		Category:    "unknown",
		Description: strings.Repeat("a", 2000),
	}

	isValid, errors := productSchema.ValidateStruct(product)
	if !isValid {
		fmt.Println("Product validation failed:")
		for field, fieldErrors := range errors {
			fmt.Printf("%s: %v\n", field, fieldErrors)
		}
	}
}

func ExampleSanitization() {
	dirtyInput := "<script>alert('xss')</script>Hello   World!\x00"
	clean := validation.SanitizeString(dirtyInput)
	fmt.Printf("Original: %q\n", dirtyInput)
	fmt.Printf("Sanitized: %q\n", clean)

	html := "<p>Hello <b>World</b></p>"
	text := validation.SanitizeHTML(html)
	fmt.Printf("HTML: %s\n", html)
	fmt.Printf("Text: %s\n", text)

	email := "  USER@EXAMPLE.COM  "
	normalized := validation.SanitizeEmail(email)
	fmt.Printf("Original email: %q\n", email)
	fmt.Printf("Normalized: %q\n", normalized)
}

func ExampleHelperFunctions() {
	// Test various helper functions
	fmt.Println("Testing helper functions:")

	// Email validation
	if validation.IsValidEmail("test@example.com") {
		fmt.Println("✓ Valid Email")
	} else {
		fmt.Println("✗ Invalid Email")
	}

	if !validation.IsValidEmail("invalid") {
		fmt.Println("✓ Correctly rejected invalid email")
	} else {
		fmt.Println("✗ Failed to reject invalid email")
	}

	// URL validation
	if validation.IsValidURL("https://example.com") {
		fmt.Println("✓ Valid URL")
	} else {
		fmt.Println("✗ Invalid URL")
	}

	// Phone validation
	if validation.IsValidPhone("+1 (555) 123-4567") {
		fmt.Println("✓ Valid Phone")
	} else {
		fmt.Println("✗ Invalid Phone")
	}

	// Credit card validation
	if validation.IsValidCreditCard("4111111111111111") {
		fmt.Println("✓ Valid Credit Card")
	} else {
		fmt.Println("✗ Invalid Credit Card")
	}

	// Domain validation
	if validation.IsValidDomain("example.com") {
		fmt.Println("✓ Valid Domain")
	} else {
		fmt.Println("✗ Invalid Domain")
	}

	// Hex color validation
	if validation.IsValidHexColor("#ff0000") {
		fmt.Println("✓ Valid Hex Color")
	} else {
		fmt.Println("✗ Invalid Hex Color")
	}

	// ISBN validation
	if validation.IsValidISBN("978-3-16-148410-0") {
		fmt.Println("✓ Valid ISBN")
	} else {
		fmt.Println("✗ Invalid ISBN")
	}
}
