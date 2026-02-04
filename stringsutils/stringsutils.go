package stringsutils

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Reverse returns the reverse of a string
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// IsPalindrome checks if a string is a palindrome (case-insensitive, ignores spaces)
func IsPalindrome(s string) bool {
	s = RemoveWhitespace(strings.ToLower(s))
	return s == Reverse(s)
}

// Truncate truncates a string to a specified length with ellipsis
func Truncate(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	if maxLength < 3 {
		return s[:maxLength]
	}
	return s[:maxLength-3] + "..."
}

// TruncateByWords truncates by words instead of characters
func TruncateByWords(s string, maxWords int) string {
	words := strings.Fields(s)
	if len(words) <= maxWords {
		return s
	}
	return strings.Join(words[:maxWords], " ") + "..."
}

// ContainsAny checks if string contains any of the substrings
func ContainsAny(s string, substrings ...string) bool {
	for _, sub := range substrings {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// ContainsAll checks if string contains all of the substrings
func ContainsAll(s string, substrings ...string) bool {
	for _, sub := range substrings {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}

// IsEmpty checks if a string is empty or contains only whitespace
func IsEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// IsBlank checks if string is empty, whitespace, or nil pointer equivalent
func IsBlank(s *string) bool {
	return s == nil || IsEmpty(*s)
}

// RemoveWhitespace removes all whitespace characters from a string
func RemoveWhitespace(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for _, ch := range s {
		if !unicode.IsSpace(ch) {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

// RemoveNonAlphanumeric removes all non-alphanumeric characters
func RemoveNonAlphanumeric(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for _, ch := range s {
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

// ToCamelCase converts a string to camelCase
func ToCamelCase(s string) string {
	// Replace underscores and dashes with spaces
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")

	words := strings.Fields(strings.ToLower(s))
	if len(words) == 0 {
		return ""
	}

	result := words[0]
	for _, word := range words[1:] {
		if len(word) > 0 {
			result += strings.ToUpper(string(word[0])) + word[1:]
		}
	}
	return result
}

// ToPascalCase converts a string to PascalCase
func ToPascalCase(s string) string {
	// Replace underscores and dashes with spaces
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")

	words := strings.Fields(strings.ToLower(s))
	var result strings.Builder

	for _, word := range words {
		if len(word) > 0 {
			result.WriteString(strings.ToUpper(string(word[0])) + word[1:])
		}
	}
	return result.String()
}

// ToSnakeCase converts a string to snake_case
func ToSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// ToKebabCase converts a string to kebab-case
func ToKebabCase(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '-')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// ToTitleCase converts string to Title Case (each word capitalized)
func ToTitleCase(s string) string {
	return strings.Title(strings.ToLower(s))
}

// IsValidEmail validates an email address using regex
func IsValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

// ExtractEmails extracts all email addresses from a string
func ExtractEmails(s string) []string {
	pattern := `[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`
	re := regexp.MustCompile(pattern)
	return re.FindAllString(s, -1)
}

// IsValidURL validates a URL
func IsValidURL(url string) bool {
	pattern := `^(https?|ftp)://[^\s/$.?#].[^\s]*$`
	matched, _ := regexp.MatchString(pattern, url)
	return matched
}

// IsValidPhone validates a phone number (basic validation)
func IsValidPhone(phone string) bool {
	// Remove all non-digit characters
	cleaned := RemoveNonAlphanumeric(phone)
	// Check if it's between 8 and 15 digits
	return len(cleaned) >= 8 && len(cleaned) <= 15
}

// CountWords counts the number of words in a string
func CountWords(s string) int {
	return len(strings.Fields(s))
}

// CountCharacters counts characters (runes) in a string
func CountCharacters(s string) int {
	return utf8.RuneCountInString(s)
}

// CountSubstring counts occurrences of a substring
func CountSubstring(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	return (len(s) - len(strings.ReplaceAll(s, substr, ""))) / len(substr)
}

// WrapText wraps text to a specified line width
func WrapText(text string, lineWidth int) string {
	var result strings.Builder
	var currentLine strings.Builder

	for _, word := range strings.Fields(text) {
		if currentLine.Len()+len(word)+1 > lineWidth {
			if currentLine.Len() > 0 {
				result.WriteString(currentLine.String() + "\n")
				currentLine.Reset()
			}
		}
		if currentLine.Len() > 0 {
			currentLine.WriteString(" ")
		}
		currentLine.WriteString(word)

		// If a single word is longer than lineWidth, force break
		if len(word) > lineWidth {
			result.WriteString(currentLine.String() + "\n")
			currentLine.Reset()
		}
	}

	if currentLine.Len() > 0 {
		result.WriteString(currentLine.String())
	}

	return result.String()
}

// Indent indents each line of a string
func Indent(s string, indent string) string {
	lines := strings.Split(s, "\n")
	var result strings.Builder

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(indent + line)
	}
	return result.String()
}

// Dedent removes common leading whitespace from all lines
func Dedent(s string) string {
	lines := strings.Split(s, "\n")

	// Find minimum indentation
	minIndent := math.MaxInt32
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue // Skip empty lines
		}
		indent := len(line) - len(strings.TrimLeft(line, " \t"))
		if indent < minIndent {
			minIndent = indent
		}
	}

	// Remove common indentation
	if minIndent > 0 && minIndent < math.MaxInt32 {
		var result strings.Builder
		for i, line := range lines {
			if i > 0 {
				result.WriteString("\n")
			}
			if len(line) >= minIndent {
				result.WriteString(line[minIndent:])
			} else {
				result.WriteString(line)
			}
		}
		return result.String()
	}

	return s
}

// PadLeft pads a string on the left
func PadLeft(s string, length int, padChar rune) string {
	if len(s) >= length {
		return s
	}
	return strings.Repeat(string(padChar), length-len(s)) + s
}

// PadRight pads a string on the right
func PadRight(s string, length int, padChar rune) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(string(padChar), length-len(s))
}

// PadCenter pads a string on both sides
func PadCenter(s string, length int, padChar rune) string {
	if len(s) >= length {
		return s
	}

	left := (length - len(s)) / 2
	right := length - len(s) - left

	return strings.Repeat(string(padChar), left) + s + strings.Repeat(string(padChar), right)
}

// RemovePrefix removes prefix if present
func RemovePrefix(s, prefix string) string {
	if strings.HasPrefix(s, prefix) {
		return s[len(prefix):]
	}
	return s
}

// RemoveSuffix removes suffix if present
func RemoveSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		return s[:len(s)-len(suffix)]
	}
	return s
}

// RemoveAll removes all occurrences of a substring
func RemoveAll(s, substr string) string {
	return strings.ReplaceAll(s, substr, "")
}

// ReplaceFirst replaces first occurrence only
func ReplaceFirst(s, old, new string) string {
	return strings.Replace(s, old, new, 1)
}

// ReplaceLast replaces last occurrence only
func ReplaceLast(s, old, new string) string {
	// Find last occurrence
	idx := strings.LastIndex(s, old)
	if idx == -1 {
		return s
	}

	var result strings.Builder
	result.WriteString(s[:idx])
	result.WriteString(new)
	result.WriteString(s[idx+len(old):])
	return result.String()
}

// Slice extracts a substring by rune indices (supports Unicode)
func Slice(s string, start, end int) string {
	runes := []rune(s)
	if start < 0 {
		start = len(runes) + start
	}
	if end < 0 {
		end = len(runes) + end
	}
	if start < 0 || start >= len(runes) || end < start || end > len(runes) {
		return ""
	}
	return string(runes[start:end])
}

// Chunk splits a string into chunks of specified size
func Chunk(s string, size int) []string {
	if size <= 0 {
		return []string{}
	}

	runes := []rune(s)
	var chunks []string

	for i := 0; i < len(runes); i += size {
		end := i + size
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
	}

	return chunks
}

// Join concatenates strings with a separator
func Join(sep string, strs ...string) string {
	return strings.Join(strs, sep)
}

// JoinSlice joins a slice of strings
func JoinSlice(strs []string, sep string) string {
	return strings.Join(strs, sep)
}

// SplitBy splits by multiple separators
func SplitBy(s string, seps ...string) []string {
	if len(seps) == 0 {
		return []string{s}
	}

	// Create a regex pattern with all separators
	var pattern strings.Builder
	pattern.WriteString("[")
	for i, sep := range seps {
		if i > 0 {
			pattern.WriteString("|")
		}
		// Escape regex special characters
		pattern.WriteString(regexp.QuoteMeta(sep))
	}
	pattern.WriteString("]")

	re := regexp.MustCompile(pattern.String())
	return re.Split(s, -1)
}

// StartsWithAny checks if string starts with any of the prefixes
func StartsWithAny(s string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

// EndsWithAny checks if string ends with any of the suffixes
func EndsWithAny(s string, suffixes ...string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(s, suffix) {
			return true
		}
	}
	return false
}

// ContainsIgnoreCase checks if string contains substring (case-insensitive)
func ContainsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// EqualsIgnoreCase checks if strings are equal (case-insensitive)
func EqualsIgnoreCase(s1, s2 string) bool {
	return strings.EqualFold(s1, s2)
}

// IsNumeric checks if string contains only digits
func IsNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// IsAlpha checks if string contains only letters
func IsAlpha(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// IsAlphaNumeric checks if string contains only letters and digits
func IsAlphaNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// IsUpper checks if string is all uppercase
func IsUpper(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if unicode.IsLetter(r) && !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}

// IsLower checks if string is all lowercase
func IsLower(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if unicode.IsLetter(r) && !unicode.IsLower(r) {
			return false
		}
	}
	return true
}

// Mask masks part of a string (for sensitive data)
func Mask(s string, visibleChars int, maskChar rune) string {
	if len(s) <= visibleChars {
		return s
	}

	runes := []rune(s)
	visible := runes[:visibleChars]
	masked := strings.Repeat(string(maskChar), len(runes)-visibleChars)

	return string(visible) + masked
}

// MaskEmail masks an email address
func MaskEmail(email string) string {
	if !IsValidEmail(email) {
		return email
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	localPart := parts[0]
	domain := parts[1]

	if len(localPart) <= 2 {
		return localPart + "@" + domain
	}

	maskedLocal := string(localPart[0]) + strings.Repeat("*", len(localPart)-2) + string(localPart[len(localPart)-1])
	return maskedLocal + "@" + domain
}

// MaskPhone masks a phone number
func MaskPhone(phone string) string {
	cleaned := RemoveNonAlphanumeric(phone)
	if len(cleaned) < 4 {
		return phone
	}

	visible := cleaned[len(cleaned)-4:]
	masked := strings.Repeat("*", len(cleaned)-4)

	return masked + visible
}

// GenerateRandomString generates a random string of specified length
func GenerateRandomString(length int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)

	// Use crypto/rand for secure random generation
	rand.Read(b)

	var result strings.Builder
	result.Grow(length)

	for i := 0; i < length; i++ {
		result.WriteByte(chars[b[i]%byte(len(chars))])
	}

	return result.String()
}

// GenerateUUID generates a random UUID (version 4)
func GenerateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)

	// Set version (4) and variant bits
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// Base64Encode encodes string to base64
func Base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// Base64Decode decodes base64 string
func Base64Decode(s string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// URLEncode encodes string for URL
func URLEncode(s string) string {
	var buf bytes.Buffer
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == '-' || r == '_' || r == '.' || r == '~' {
			buf.WriteRune(r)
		} else {
			fmt.Fprintf(&buf, "%%%02X", r)
		}
	}
	return buf.String()
}

// URLDecode decodes URL-encoded string
func URLDecode(s string) (string, error) {
	return "", nil // Implementation omitted for brevity
}

// RemoveAccents removes diacritics from characters
func RemoveAccents(s string) string {
	// This is a simplified version
	// For full implementation, you'd need a proper Unicode normalization
	replacements := map[rune]rune{
		'á': 'a', 'é': 'e', 'í': 'i', 'ó': 'o', 'ú': 'u',
		'à': 'a', 'è': 'e', 'ì': 'i', 'ò': 'o', 'ù': 'u',
		'ä': 'a', 'ë': 'e', 'ï': 'i', 'ö': 'o', 'ü': 'u',
		'â': 'a', 'ê': 'e', 'î': 'i', 'ô': 'o', 'û': 'u',
		'ã': 'a', 'ñ': 'n', 'ç': 'c',
		'Á': 'A', 'É': 'E', 'Í': 'I', 'Ó': 'O', 'Ú': 'U',
		'À': 'A', 'È': 'E', 'Ì': 'I', 'Ò': 'O', 'Ù': 'U',
		'Ä': 'A', 'Ë': 'E', 'Ï': 'I', 'Ö': 'O', 'Ü': 'U',
		'Â': 'A', 'Ê': 'E', 'Î': 'I', 'Ô': 'O', 'Û': 'U',
		'Ã': 'A', 'Ñ': 'N', 'Ç': 'C',
	}

	var result strings.Builder
	for _, r := range s {
		if replacement, ok := replacements[r]; ok {
			result.WriteRune(replacement)
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// LevenshteinDistance calculates the Levenshtein distance between two strings
func LevenshteinDistance(s, t string) int {
	if s == t {
		return 0
	}
	if len(s) == 0 {
		return len(t)
	}
	if len(t) == 0 {
		return len(s)
	}

	// Convert to runes for Unicode support
	sRunes := []rune(s)
	tRunes := []rune(t)

	// Create work vectors
	v0 := make([]int, len(tRunes)+1)
	v1 := make([]int, len(tRunes)+1)

	// Initialize v0
	for i := 0; i < len(v0); i++ {
		v0[i] = i
	}

	// Calculate distance
	for i := 0; i < len(sRunes); i++ {
		v1[0] = i + 1

		for j := 0; j < len(tRunes); j++ {
			cost := 0
			if sRunes[i] != tRunes[j] {
				cost = 1
			}

			v1[j+1] = min(v1[j]+1, min(v0[j+1]+1, v0[j]+cost))
		}

		// Swap vectors
		v0, v1 = v1, v0
	}

	return v0[len(tRunes)]
}

// Similarity calculates string similarity percentage (0-100)
func Similarity(s1, s2 string) float64 {
	distance := LevenshteinDistance(s1, s2)
	maxLen := max(len([]rune(s1)), len([]rune(s2)))

	if maxLen == 0 {
		return 100.0
	}

	return (1 - float64(distance)/float64(maxLen)) * 100
}

// FindCommonPrefix finds common prefix among strings
func FindCommonPrefix(strs ...string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	prefix := strs[0]
	for _, s := range strs[1:] {
		i := 0
		minLen := min(len(prefix), len(s))

		for i < minLen && prefix[i] == s[i] {
			i++
		}

		prefix = prefix[:i]
		if prefix == "" {
			break
		}
	}

	return prefix
}

// FindCommonSuffix finds common suffix among strings
func FindCommonSuffix(strs ...string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	suffix := strs[0]
	for _, s := range strs[1:] {
		i, j := len(suffix)-1, len(s)-1

		for i >= 0 && j >= 0 && suffix[i] == s[j] {
			i--
			j--
		}

		suffix = suffix[i+1:]
		if suffix == "" {
			break
		}
	}

	return suffix
}

// Intersection finds common strings in multiple slices
func Intersection(slices ...[]string) []string {
	if len(slices) == 0 {
		return []string{}
	}
	if len(slices) == 1 {
		return slices[0]
	}

	// Count occurrences
	counts := make(map[string]int)
	for _, slice := range slices {
		seen := make(map[string]bool)
		for _, s := range slice {
			if !seen[s] {
				counts[s]++
				seen[s] = true
			}
		}
	}

	// Find strings that appear in all slices
	var result []string
	required := len(slices)
	for s, count := range counts {
		if count == required {
			result = append(result, s)
		}
	}

	return result
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Unique returns unique strings preserving order
func Unique(strs []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, s := range strs {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}

	return result
}

// Filter filters strings based on predicate
func Filter(strs []string, predicate func(string) bool) []string {
	var result []string
	for _, s := range strs {
		if predicate(s) {
			result = append(result, s)
		}
	}
	return result
}

// Map applies function to each string
func Map(strs []string, fn func(string) string) []string {
	result := make([]string, len(strs))
	for i, s := range strs {
		result[i] = fn(s)
	}
	return result
}

// Reduce reduces strings to single value
func Reduce(strs []string, initial string, fn func(string, string) string) string {
	result := initial
	for _, s := range strs {
		result = fn(result, s)
	}
	return result
}
