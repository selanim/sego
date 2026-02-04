package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// Validator represents a validation instance
type Validator struct {
	errors map[string][]string
}

// New creates a new validator instance
func New() *Validator {
	return &Validator{
		errors: make(map[string][]string),
	}
}

// Error represents a validation error
type Error struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Errors represents multiple validation errors
type Errors struct {
	Errors []Error `json:"errors"`
}

// Error implements error interface
func (e Errors) Error() string {
	var msgs []string
	for _, err := range e.Errors {
		msgs = append(msgs, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(msgs, "; ")
}

// Rules defines validation rules for a field
type Rules struct {
	Required     bool
	Min          *int
	Max          *int
	MinFloat     *float64 // Add float support
	MaxFloat     *float64 // Add float support
	MinLen       *int
	MaxLen       *int
	Pattern      *regexp.Regexp
	Email        bool
	URL          bool
	Alpha        bool
	AlphaNum     bool
	Numeric      bool
	UUID         bool
	IP           bool
	IPv4         bool
	IPv6         bool
	OneOf        []interface{}
	NotOneOf     []interface{}
	Custom       []func(interface{}) error
	CustomMsg    string
	TimeFormat   *string
	TimeAfter    *time.Time
	TimeBefore   *time.Time
	DateTimeOnly bool
	DateOnly     bool
	TimeOnly     bool
}

// Validate validates a struct based on field tags
func (v *Validator) Validate(s interface{}) bool {
	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)
		fieldName := field.Name
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			fieldName = strings.Split(jsonTag, ",")[0]
		}

		// Get validation tag
		validateTag := field.Tag.Get("validate")
		if validateTag != "" {
			v.validateField(fieldName, value.Interface(), validateTag)
		}
	}

	return len(v.errors) == 0
}

// ValidateField validates a single field with rules
func (v *Validator) ValidateField(field string, value interface{}, rules string) bool {
	v.validateField(field, value, rules)
	return len(v.errors[field]) == 0
}

// AddError adds a custom error for a field
func (v *Validator) AddError(field, message string) {
	v.errors[field] = append(v.errors[field], message)
}

// HasErrors checks if there are any validation errors
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// GetErrors returns all validation errors
func (v *Validator) GetErrors() map[string][]string {
	return v.errors
}

// GetErrorMap returns errors as a map
func (v *Validator) GetErrorMap() map[string]interface{} {
	errMap := make(map[string]interface{})
	for field, errors := range v.errors {
		errMap[field] = errors
	}
	return errMap
}

// GetErrorList returns errors as a list
func (v *Validator) GetErrorList() []Error {
	var errors []Error
	for field, messages := range v.errors {
		for _, message := range messages {
			errors = append(errors, Error{Field: field, Message: message})
		}
	}
	return errors
}

// Clear clears all validation errors
func (v *Validator) Clear() {
	v.errors = make(map[string][]string)
}

// validateField validates a field based on validation tag
func (v *Validator) validateField(field string, value interface{}, rules string) {
	ruleList := strings.Split(rules, "|")
	for _, rule := range ruleList {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}

		parts := strings.SplitN(rule, ":", 2)
		ruleName := parts[0]
		var ruleValue string
		if len(parts) > 1 {
			ruleValue = parts[1]
		}

		switch ruleName {
		case "required":
			if isEmpty(value) {
				v.AddError(field, "This field is required")
			}
		case "min":
			if num, ok := toInt(ruleValue); ok {
				if intVal, err := toIntValue(value); err == nil && intVal < num {
					v.AddError(field, fmt.Sprintf("Must be at least %d", num))
				}
			}
		case "max":
			if num, ok := toInt(ruleValue); ok {
				if intVal, err := toIntValue(value); err == nil && intVal > num {
					v.AddError(field, fmt.Sprintf("Must be at most %d", num))
				}
			}
		case "min_len":
			if length, ok := toInt(ruleValue); ok {
				if str, isString := toString(value); isString && utf8.RuneCountInString(str) < length {
					v.AddError(field, fmt.Sprintf("Must be at least %d characters", length))
				}
			}
		case "max_len":
			if length, ok := toInt(ruleValue); ok {
				if str, isString := toString(value); isString && utf8.RuneCountInString(str) > length {
					v.AddError(field, fmt.Sprintf("Must be at most %d characters", length))
				}
			}
		case "email":
			if str, isString := toString(value); isString && str != "" {
				if !isEmail(str) {
					v.AddError(field, "Must be a valid email address")
				}
			}
		case "url":
			if str, isString := toString(value); isString && str != "" {
				if !isURL(str) {
					v.AddError(field, "Must be a valid URL")
				}
			}
		case "alpha":
			if str, isString := toString(value); isString && str != "" {
				if !isAlpha(str) {
					v.AddError(field, "Must contain only letters")
				}
			}
		case "alphanum":
			if str, isString := toString(value); isString && str != "" {
				if !isAlphaNum(str) {
					v.AddError(field, "Must contain only letters and numbers")
				}
			}
		case "numeric":
			if str, isString := toString(value); isString && str != "" {
				if !isNumeric(str) {
					v.AddError(field, "Must be a valid number")
				}
			}
		case "uuid":
			if str, isString := toString(value); isString && str != "" {
				if !isUUID(str) {
					v.AddError(field, "Must be a valid UUID")
				}
			}
		case "ip":
			if str, isString := toString(value); isString && str != "" {
				if !isIP(str) {
					v.AddError(field, "Must be a valid IP address")
				}
			}
		case "ipv4":
			if str, isString := toString(value); isString && str != "" {
				if !isIPv4(str) {
					v.AddError(field, "Must be a valid IPv4 address")
				}
			}
		case "ipv6":
			if str, isString := toString(value); isString && str != "" {
				if !isIPv6(str) {
					v.AddError(field, "Must be a valid IPv6 address")
				}
			}
		case "regex":
			if pattern, err := regexp.Compile(ruleValue); err == nil {
				if str, isString := toString(value); isString && str != "" {
					if !pattern.MatchString(str) {
						v.AddError(field, "Does not match required pattern")
					}
				}
			}
		case "in":
			if str, isString := toString(value); isString && str != "" {
				options := strings.Split(ruleValue, ",")
				found := false
				for _, opt := range options {
					if strings.TrimSpace(opt) == str {
						found = true
						break
					}
				}
				if !found {
					v.AddError(field, fmt.Sprintf("Must be one of: %s", ruleValue))
				}
			}
		case "not_in":
			if str, isString := toString(value); isString && str != "" {
				options := strings.Split(ruleValue, ",")
				for _, opt := range options {
					if strings.TrimSpace(opt) == str {
						v.AddError(field, fmt.Sprintf("Must not be: %s", str))
						break
					}
				}
			}
		case "date":
			if str, isString := toString(value); isString && str != "" {
				if _, err := time.Parse("2006-01-02", str); err != nil {
					v.AddError(field, "Must be a valid date (YYYY-MM-DD)")
				}
			}
		case "datetime":
			if str, isString := toString(value); isString && str != "" {
				if _, err := time.Parse(time.RFC3339, str); err != nil {
					v.AddError(field, "Must be a valid datetime (RFC3339)")
				}
			}
		case "time":
			if str, isString := toString(value); isString && str != "" {
				if _, err := time.Parse("15:04:05", str); err != nil {
					v.AddError(field, "Must be a valid time (HH:MM:SS)")
				}
			}
		case "equal":
			if str, isString := toString(value); isString && str != ruleValue {
				v.AddError(field, fmt.Sprintf("Must be equal to %s", ruleValue))
			}
		case "not_equal":
			if str, isString := toString(value); isString && str == ruleValue {
				v.AddError(field, fmt.Sprintf("Must not be equal to %s", ruleValue))
			}
		}
	}
}

// Helper functions
func isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
		return v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	}

	// Check zero value for other types
	return reflect.DeepEqual(value, reflect.Zero(reflect.TypeOf(value)).Interface())
}

func toInt(s string) (int, bool) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err == nil
}

func toIntValue(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int8:
		return int(v), nil
	case int16:
		return int(v), nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case uint:
		return int(v), nil
	case uint8:
		return int(v), nil
	case uint16:
		return int(v), nil
	case uint32:
		return int(v), nil
	case uint64:
		return int(v), nil
	case float32:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		var n int
		_, err := fmt.Sscanf(v, "%d", &n)
		return n, err
	default:
		return 0, fmt.Errorf("cannot convert to int")
	}
}

func toString(value interface{}) (string, bool) {
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

// Validation patterns
var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	urlRegex      = regexp.MustCompile(`^(https?://)?([a-zA-Z0-9.-]+\.[a-zA-Z]{2,})(:[0-9]+)?(/.*)?$`)
	alphaRegex    = regexp.MustCompile(`^[a-zA-Z]+$`)
	alphaNumRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	numericRegex  = regexp.MustCompile(`^-?\d+(\.\d+)?$`)
	uuidRegex     = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	ipRegex       = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$|^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$`)
	ipv4Regex     = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	ipv6Regex     = regexp.MustCompile(`^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$`)
)

func isEmail(s string) bool {
	return emailRegex.MatchString(s)
}

func isURL(s string) bool {
	return urlRegex.MatchString(s)
}

func isAlpha(s string) bool {
	return alphaRegex.MatchString(s)
}

func isAlphaNum(s string) bool {
	return alphaNumRegex.MatchString(s)
}

func isNumeric(s string) bool {
	return numericRegex.MatchString(s)
}

func isUUID(s string) bool {
	return uuidRegex.MatchString(strings.ToLower(s))
}

func isIP(s string) bool {
	return ipv4Regex.MatchString(s) || ipv6Regex.MatchString(s)
}

func isIPv4(s string) bool {
	return ipv4Regex.MatchString(s)
}

func isIPv6(s string) bool {
	return ipv6Regex.MatchString(s)
}
