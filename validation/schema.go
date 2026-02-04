package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// Schema defines validation rules for a type
type Schema struct {
	rules map[string]Rules
}

// NewSchema creates a new validation schema
func NewSchema() *Schema {
	return &Schema{
		rules: make(map[string]Rules),
	}
}

// Field adds validation rules for a field
func (s *Schema) Field(field string, rules ...func(*Rules)) *Schema {
	r := Rules{}
	for _, ruleFunc := range rules {
		ruleFunc(&r)
	}
	s.rules[field] = r
	return s
}

// Rule helper functions
func Required() func(*Rules) {
	return func(r *Rules) { r.Required = true }
}

func Min(min int) func(*Rules) {
	return func(r *Rules) { r.Min = &min }
}

func Max(max int) func(*Rules) {
	return func(r *Rules) { r.Max = &max }
}

func MinLen(min int) func(*Rules) {
	return func(r *Rules) { r.MinLen = &min }
}

func MaxLen(max int) func(*Rules) {
	return func(r *Rules) { r.MaxLen = &max }
}

func Pattern(pattern string) func(*Rules) {
	return func(r *Rules) {
		if re, err := regexp.Compile(pattern); err == nil {
			r.Pattern = re
		}
	}
}

func Email() func(*Rules) {
	return func(r *Rules) { r.Email = true }
}

func URL() func(*Rules) {
	return func(r *Rules) { r.URL = true }
}

func Alpha() func(*Rules) {
	return func(r *Rules) { r.Alpha = true }
}

func AlphaNum() func(*Rules) {
	return func(r *Rules) { r.AlphaNum = true }
}

func Numeric() func(*Rules) {
	return func(r *Rules) { r.Numeric = true }
}

func UUID() func(*Rules) {
	return func(r *Rules) { r.UUID = true }
}

func IP() func(*Rules) {
	return func(r *Rules) { r.IP = true }
}

func IPv4() func(*Rules) {
	return func(r *Rules) { r.IPv4 = true }
}

func IPv6() func(*Rules) {
	return func(r *Rules) { r.IPv6 = true }
}

func OneOf(values ...interface{}) func(*Rules) {
	return func(r *Rules) { r.OneOf = values }
}

func NotOneOf(values ...interface{}) func(*Rules) {
	return func(r *Rules) { r.NotOneOf = values }
}

func Custom(fn func(interface{}) error, msg ...string) func(*Rules) {
	return func(r *Rules) {
		r.Custom = append(r.Custom, fn)
		if len(msg) > 0 {
			r.CustomMsg = msg[0]
		}
	}
}

func TimeFormat(format string) func(*Rules) {
	return func(r *Rules) { r.TimeFormat = &format }
}

func TimeAfter(t time.Time) func(*Rules) {
	return func(r *Rules) { r.TimeAfter = &t }
}

func TimeBefore(t time.Time) func(*Rules) {
	return func(r *Rules) { r.TimeBefore = &t }
}

func DateTimeOnly() func(*Rules) {
	return func(r *Rules) { r.DateTimeOnly = true }
}

func DateOnly() func(*Rules) {
	return func(r *Rules) { r.DateOnly = true }
}

func TimeOnly() func(*Rules) {
	return func(r *Rules) { r.TimeOnly = true }
}

// Validate validates data against the schema
func (s *Schema) Validate(data interface{}) (bool, map[string][]string) {
	validator := New()

	// If data is a map
	if m, ok := data.(map[string]interface{}); ok {
		for field, rules := range s.rules {
			if value, exists := m[field]; exists {
				s.validateField(validator, field, value, rules)
			} else if rules.Required {
				validator.AddError(field, "This field is required")
			}
		}
	} else {
		// If data is a struct
		val := reflect.ValueOf(data)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		typ := val.Type()
		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			value := val.Field(i).Interface()
			fieldName := field.Name

			// Check for json tag
			if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
				fieldName = strings.Split(jsonTag, ",")[0]
			}

			if rules, exists := s.rules[fieldName]; exists {
				s.validateField(validator, fieldName, value, rules)
			}
		}
	}

	return !validator.HasErrors(), validator.GetErrors()
}

// validateField validates a single field against rules
func (s *Schema) validateField(v *Validator, field string, value interface{}, rules Rules) {
	// Check required
	if rules.Required && isEmpty(value) {
		v.AddError(field, "This field is required")
		return
	}

	// Skip further validation if field is empty and not required
	if isEmpty(value) {
		return
	}

	// Min/Max for numbers
	if rules.Min != nil {
		if intVal, err := toIntValue(value); err == nil && intVal < *rules.Min {
			v.AddError(field, fmt.Sprintf("Must be at least %d", *rules.Min))
		}
	}

	if rules.Max != nil {
		if intVal, err := toIntValue(value); err == nil && intVal > *rules.Max {
			v.AddError(field, fmt.Sprintf("Must be at most %d", *rules.Max))
		}
	}

	// MinLen/MaxLen for strings
	if str, isString := toString(value); isString {
		length := utf8.RuneCountInString(str)

		if rules.MinLen != nil && length < *rules.MinLen {
			v.AddError(field, fmt.Sprintf("Must be at least %d characters", *rules.MinLen))
		}

		if rules.MaxLen != nil && length > *rules.MaxLen {
			v.AddError(field, fmt.Sprintf("Must be at most %d characters", *rules.MaxLen))
		}

		// Pattern matching
		if rules.Pattern != nil && !rules.Pattern.MatchString(str) {
			v.AddError(field, "Does not match required pattern")
		}

		// Email validation
		if rules.Email && !isEmail(str) {
			v.AddError(field, "Must be a valid email address")
		}

		// URL validation
		if rules.URL && !isURL(str) {
			v.AddError(field, "Must be a valid URL")
		}

		// Alpha validation
		if rules.Alpha && !isAlpha(str) {
			v.AddError(field, "Must contain only letters")
		}

		// AlphaNum validation
		if rules.AlphaNum && !isAlphaNum(str) {
			v.AddError(field, "Must contain only letters and numbers")
		}

		// Numeric validation
		if rules.Numeric && !isNumeric(str) {
			v.AddError(field, "Must be a valid number")
		}

		// UUID validation
		if rules.UUID && !isUUID(str) {
			v.AddError(field, "Must be a valid UUID")
		}

		// IP validation
		if rules.IP && !isIP(str) {
			v.AddError(field, "Must be a valid IP address")
		}

		// IPv4 validation
		if rules.IPv4 && !isIPv4(str) {
			v.AddError(field, "Must be a valid IPv4 address")
		}

		// IPv6 validation
		if rules.IPv6 && !isIPv6(str) {
			v.AddError(field, "Must be a valid IPv6 address")
		}

		// OneOf validation
		if len(rules.OneOf) > 0 {
			found := false
			for _, opt := range rules.OneOf {
				if fmt.Sprintf("%v", opt) == str {
					found = true
					break
				}
			}
			if !found {
				var options []string
				for _, opt := range rules.OneOf {
					options = append(options, fmt.Sprintf("%v", opt))
				}
				v.AddError(field, fmt.Sprintf("Must be one of: %s", strings.Join(options, ", ")))
			}
		}

		// NotOneOf validation
		if len(rules.NotOneOf) > 0 {
			for _, opt := range rules.NotOneOf {
				if fmt.Sprintf("%v", opt) == str {
					v.AddError(field, fmt.Sprintf("Must not be: %v", opt))
					break
				}
			}
		}

		// Time validation
		if rules.TimeFormat != nil {
			if _, err := time.Parse(*rules.TimeFormat, str); err != nil {
				v.AddError(field, fmt.Sprintf("Must be a valid date/time in format: %s", *rules.TimeFormat))
			}
		}

		if rules.DateTimeOnly {
			if _, err := time.Parse(time.RFC3339, str); err != nil {
				v.AddError(field, "Must be a valid datetime (RFC3339)")
			}
		}

		if rules.DateOnly {
			if _, err := time.Parse("2006-01-02", str); err != nil {
				v.AddError(field, "Must be a valid date (YYYY-MM-DD)")
			}
		}

		if rules.TimeOnly {
			if _, err := time.Parse("15:04:05", str); err != nil {
				v.AddError(field, "Must be a valid time (HH:MM:SS)")
			}
		}
	}

	// Time validation for time.Time values
	if t, ok := value.(time.Time); ok {
		if rules.TimeAfter != nil && !t.After(*rules.TimeAfter) {
			v.AddError(field, fmt.Sprintf("Must be after %s", rules.TimeAfter.Format(time.RFC3339)))
		}

		if rules.TimeBefore != nil && !t.Before(*rules.TimeBefore) {
			v.AddError(field, fmt.Sprintf("Must be before %s", rules.TimeBefore.Format(time.RFC3339)))
		}
	}

	// Custom validation functions
	for _, customFn := range rules.Custom {
		if err := customFn(value); err != nil {
			if rules.CustomMsg != "" {
				v.AddError(field, rules.CustomMsg)
			} else {
				v.AddError(field, err.Error())
			}
		}
	}
}

// ValidateStruct validates a struct against the schema
func (s *Schema) ValidateStruct(data interface{}) (bool, map[string][]string) {
	return s.Validate(data)
}

// ValidateMap validates a map against the schema
func (s *Schema) ValidateMap(data map[string]interface{}) (bool, map[string][]string) {
	return s.Validate(data)
}
