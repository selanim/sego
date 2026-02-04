package stringsutils

import (
	"fmt"
	"strings"
	"testing"
)

// TestReverse tests Reverse function
func TestReverse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"single character", "a", "a"},
		{"hello world", "hello", "olleh"},
		{"unicode", "café", "éfac"},
		{"emoji", "👋🌍", "🌍👋"},
		{"palindrome", "radar", "radar"},
		{"with spaces", "hello world", "dlrow olleh"},
		{"numbers", "12345", "54321"},
		{"mixed", "a1b2c3", "3c2b1a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Reverse(tt.input)
			if result != tt.expected {
				t.Errorf("Reverse(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestIsPalindrome tests IsPalindrome function
func TestIsPalindrome(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"empty", "", true},
		{"single char", "a", true},
		{"palindrome odd", "radar", true},
		{"palindrome even", "abba", true},
		{"not palindrome", "hello", false},
		{"case insensitive", "Racecar", true},
		{"with spaces", "A man a plan a canal Panama", true},
		{"with punctuation", "Madam, I'm Adam", false}, // punctuation not handled
		{"numbers", "12321", true},
		{"mixed case", "RaceCar", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPalindrome(tt.input)
			if result != tt.expected {
				t.Errorf("IsPalindrome(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestTruncate tests Truncate function
func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		max      int
		expected string
	}{
		{"no truncation needed", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"truncate with ellipsis", "hello world", 8, "hello..."},
		{"very short max", "hello", 2, "he"},
		{"max less than 3", "hello", 1, "h"},
		{"empty string", "", 5, ""},
		{"unicode truncate", "café latte", 6, "caf..."},
		{"emoji truncate", "👋🌍🎉", 2, "👋🌍"}, // Note: emoji are multiple bytes
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Truncate(tt.input, tt.max)
			if result != tt.expected {
				t.Errorf("Truncate(%q, %d) = %q, expected %q", tt.input, tt.max, result, tt.expected)
			}
		})
	}
}

// TestTruncateByWords tests TruncateByWords function
func TestTruncateByWords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWords int
		expected string
	}{
		{"no truncation", "hello world", 3, "hello world"},
		{"truncate one word", "hello", 0, "..."},
		{"multiple words", "the quick brown fox", 2, "the quick..."},
		{"with extra spaces", "  hello   world  ", 1, "hello..."},
		{"empty string", "", 3, ""},
		{"exact word count", "one two three", 3, "one two three"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateByWords(tt.input, tt.maxWords)
			if result != tt.expected {
				t.Errorf("TruncateByWords(%q, %d) = %q, expected %q", tt.input, tt.maxWords, result, tt.expected)
			}
		})
	}
}

// TestContainsAny tests ContainsAny function
func TestContainsAny(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		substrs  []string
		expected bool
	}{
		{"contains one", "hello world", []string{"world"}, true},
		{"contains none", "hello world", []string{"foo", "bar"}, false},
		{"contains first", "hello world", []string{"hello", "world"}, true},
		{"contains second", "hello world", []string{"foo", "world"}, true},
		{"empty string", "", []string{"test"}, false},
		{"empty substr", "hello", []string{""}, true},
		{"multiple matches", "hello hello", []string{"hello"}, true},
		{"unicode", "café au lait", []string{"café"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsAny(tt.input, tt.substrs...)
			if result != tt.expected {
				t.Errorf("ContainsAny(%q, %v) = %v, expected %v", tt.input, tt.substrs, result, tt.expected)
			}
		})
	}
}

// TestContainsAll tests ContainsAll function
func TestContainsAll(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		substrs  []string
		expected bool
	}{
		{"contains all", "hello world foo bar", []string{"hello", "world"}, true},
		{"missing one", "hello world", []string{"hello", "foo"}, false},
		{"empty input", "", []string{"test"}, false},
		{"empty substr list", "hello", []string{}, true},
		{"case sensitive", "Hello World", []string{"hello"}, false},
		{"overlap", "hello hello", []string{"hello"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsAll(tt.input, tt.substrs...)
			if result != tt.expected {
				t.Errorf("ContainsAll(%q, %v) = %v, expected %v", tt.input, tt.substrs, result, tt.expected)
			}
		})
	}
}

// TestIsEmpty tests IsEmpty function
func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"empty", "", true},
		{"only spaces", "   ", true},
		{"tabs and spaces", "\t\n\r ", true},
		{"non-empty", "hello", false},
		{"non-empty with spaces", " hello ", false},
		{"zero width space", "\u200B", false}, // Not trimmed by TrimSpace
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEmpty(tt.input)
			if result != tt.expected {
				t.Errorf("IsEmpty(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestRemoveWhitespace tests RemoveWhitespace function
func TestRemoveWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no whitespace", "hello", "hello"},
		{"with spaces", "hello world", "helloworld"},
		{"with tabs", "hello\tworld", "helloworld"},
		{"mixed whitespace", "h e\tl\nl\to", "hello"},
		{"leading/trailing", "  hello  ", "hello"},
		{"empty", "", ""},
		{"only whitespace", " \t\n\r ", ""},
		{"unicode spaces", "hello\u200Bworld", "hello\u200Bworld"}, // zero width space not removed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveWhitespace(tt.input)
			if result != tt.expected {
				t.Errorf("RemoveWhitespace(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestToCamelCase tests ToCamelCase function
func TestToCamelCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"snake_case", "hello_world", "helloWorld"},
		{"kebab-case", "hello-world", "helloWorld"},
		{"spaces", "hello world", "helloWorld"},
		{"already camel", "helloWorld", "helloworld"},
		{"pascal case", "HelloWorld", "helloworld"},
		{"mixed", "hello_World-Foo bar", "helloWorldFooBar"},
		{"with numbers", "user_id_123", "userId123"},
		{"empty", "", ""},
		{"single word", "hello", "hello"},
		{"multiple underscores", "hello___world", "helloWorld"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToCamelCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToCamelCase(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestToPascalCase tests ToPascalCase function
func TestToPascalCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"snake_case", "hello_world", "HelloWorld"},
		{"kebab-case", "hello-world", "HelloWorld"},
		{"spaces", "hello world", "HelloWorld"},
		{"camel case", "helloWorld", "Helloworld"},
		{"already pascal", "HelloWorld", "HelloWorld"},
		{"mixed", "hello_World-Foo bar", "HelloWorldFooBar"},
		{"with numbers", "user_2_id", "User2Id"},
		{"empty", "", ""},
		{"single word", "hello", "Hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToPascalCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToPascalCase(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestToSnakeCase tests ToSnakeCase function
func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"camelCase", "helloWorld", "hello_world"},
		{"PascalCase", "HelloWorld", "hello_world"},
		{"already snake", "hello_world", "hello_world"},
		{"with acronym", "XMLHttpRequest", "x_m_l_http_request"},
		{"single word", "hello", "hello"},
		{"empty", "", ""},
		{"multiple caps", "HTTPServer", "h_t_t_p_server"},
		{"with numbers", "userID123", "user_i_d123"},
		{"mixed case", "Hello_World-Foo", "hello_world-foo"}, // dash preserved
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToSnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToSnakeCase(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestToKebabCase tests ToKebabCase function
func TestToKebabCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"camelCase", "helloWorld", "hello-world"},
		{"PascalCase", "HelloWorld", "hello-world"},
		{"already kebab", "hello-world", "hello-world"},
		{"single word", "hello", "hello"},
		{"empty", "", ""},
		{"with numbers", "userID123", "user-i-d123"},
		{"mixed case", "Hello_WorldFoo", "hello_world-foo"}, // underscore preserved
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToKebabCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToKebabCase(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestIsValidEmail tests IsValidEmail function
func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid simple", "test@example.com", true},
		{"valid with dot", "test.name@example.com", true},
		{"valid with plus", "test+tag@example.com", true},
		{"valid with dash", "test-name@example.com", true},
		{"valid subdomain", "test@sub.example.com", true},
		{"invalid no @", "testexample.com", false},
		{"invalid no domain", "test@", false},
		{"invalid no TLD", "test@example", false},
		{"invalid spaces", "test @example.com", false},
		{"invalid double @", "test@@example.com", false},
		{"valid with numbers", "test123@example.com", true},
		{"valid with underscore", "test_name@example.com", true},
		{"invalid starting dot", ".test@example.com", false},
		{"valid long TLD", "test@example.travel", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidEmail(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidEmail(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestExtractEmails tests ExtractEmails function
func TestExtractEmails(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"single email", "Contact me at test@example.com", []string{"test@example.com"}},
		{"multiple emails", "Email1: a@b.com, Email2: c@d.com", []string{"a@b.com", "c@d.com"}},
		{"no emails", "This has no email", []string{}},
		{"with punctuation", "Email: test@example.com, and another.", []string{"test@example.com"}},
		{"invalid in text", "My email is @invalid.com", []string{}},
		{"multiple lines", "a@b.com\nc@d.com\ne@f.com", []string{"a@b.com", "c@d.com", "e@f.com"}},
		{"mixed valid/invalid", "Valid: test@example.com, invalid: @example.com", []string{"test@example.com"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractEmails(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("ExtractEmails(%q) = %v, expected %v", tt.input, result, tt.expected)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("ExtractEmails(%q)[%d] = %q, expected %q", tt.input, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// TestCountWords tests CountWords function
func TestCountWords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"empty", "", 0},
		{"single word", "hello", 1},
		{"two words", "hello world", 2},
		{"with punctuation", "hello, world!", 2},
		{"multiple spaces", "hello   world", 2},
		{"tabs and newlines", "hello\tworld\nfoo", 3},
		{"leading/trailing spaces", "  hello world  ", 2},
		{"only spaces", "     ", 0},
		{"unicode words", "café au lait", 3},
		{"mixed whitespace", "hello \t\n world", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountWords(tt.input)
			if result != tt.expected {
				t.Errorf("CountWords(%q) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

// TestCountCharacters tests CountCharacters function
func TestCountCharacters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"empty", "", 0},
		{"ascii", "hello", 5},
		{"unicode", "café", 4},
		{"emoji", "👋🌍", 2},
		{"mixed", "a👋b🌍c", 5},
		{"with spaces", "hello world", 11},
		{"newline", "hello\nworld", 11}, // \n counts as 1 character
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountCharacters(tt.input)
			if result != tt.expected {
				t.Errorf("CountCharacters(%q) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

// TestCountSubstring tests CountSubstring function
func TestCountSubstring(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		substr   string
		expected int
	}{
		{"empty string", "", "test", 0},
		{"empty substring", "test", "", 0},
		{"no match", "hello", "world", 0},
		{"single match", "hello world", "world", 1},
		{"multiple matches", "hello hello hello", "hello", 3},
		{"overlapping", "aaaa", "aa", 2}, // Non-overlapping count
		{"case sensitive", "Hello hello", "hello", 1},
		{"unicode", "café café", "café", 2},
		{"partial match", "hello", "ell", 1},
		{"entire string", "hello", "hello", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountSubstring(tt.input, tt.substr)
			if result != tt.expected {
				t.Errorf("CountSubstring(%q, %q) = %d, expected %d", tt.input, tt.substr, result, tt.expected)
			}
		})
	}
}

// TestWrapText tests WrapText function
func TestWrapText(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		lineWidth int
		expected  string
	}{
		{"no wrap needed", "short", 10, "short"},
		{"simple wrap", "hello world", 5, "hello\nworld"},
		{"multiple lines", "the quick brown fox", 10, "the quick\nbrown fox"},
		{"long word", "supercalifragilisticexpialidocious", 10, "supercalifragilisticexpialidocious"},
		{"with punctuation", "hello, world! foo bar.", 8, "hello,\nworld!\nfoo bar."},
		{"preserve existing newlines", "hello\nworld", 20, "hello\nworld"},
		{"empty string", "", 10, ""},
		{"only spaces", "   ", 5, ""},
		{"tab character", "hello\tworld", 20, "hello\tworld"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapText(tt.input, tt.lineWidth)
			if result != tt.expected {
				t.Errorf("WrapText(%q, %d) = %q, expected %q", tt.input, tt.lineWidth, result, tt.expected)
			}
		})
	}
}

// TestPadLeft tests PadLeft function
func TestPadLeft(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		length   int
		padChar  rune
		expected string
	}{
		{"no padding needed", "hello", 5, '*', "hello"},
		{"pad left", "hello", 10, '*', "*****hello"},
		{"pad with space", "hello", 8, ' ', "   hello"},
		{"pad with zero", "42", 5, '0', "00042"},
		{"unicode pad char", "test", 6, '★', "★★test"},
		{"empty string", "", 5, '*', "*****"},
		{"negative length", "hello", -1, '*', "hello"},
		{"zero length", "hello", 0, '*', "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PadLeft(tt.input, tt.length, tt.padChar)
			if result != tt.expected {
				t.Errorf("PadLeft(%q, %d, %q) = %q, expected %q", tt.input, tt.length, tt.padChar, result, tt.expected)
			}
		})
	}
}

// TestPadRight tests PadRight function
func TestPadRight(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		length   int
		padChar  rune
		expected string
	}{
		{"no padding needed", "hello", 5, '*', "hello"},
		{"pad right", "hello", 10, '*', "hello*****"},
		{"pad with space", "hello", 8, ' ', "hello   "},
		{"pad with dash", "test", 6, '-', "test--"},
		{"empty string", "", 5, '*', "*****"},
		{"unicode string", "café", 6, ' ', "café  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PadRight(tt.input, tt.length, tt.padChar)
			if result != tt.expected {
				t.Errorf("PadRight(%q, %d, %q) = %q, expected %q", tt.input, tt.length, tt.padChar, result, tt.expected)
			}
		})
	}
}

// TestPadCenter tests PadCenter function
func TestPadCenter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		length   int
		padChar  rune
		expected string
	}{
		{"no padding needed", "hello", 5, '*', "hello"},
		{"even padding", "hi", 6, '*', "**hi**"},
		{"odd padding", "hi", 5, '*', "*hi**"},
		{"odd string length", "hello", 8, '*', "*hello**"},
		{"pad with space", "test", 10, ' ', "   test   "},
		{"empty string", "", 5, '*', "*****"},
		{"unicode", "café", 8, ' ', "  café  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PadCenter(tt.input, tt.length, tt.padChar)
			if result != tt.expected {
				t.Errorf("PadCenter(%q, %d, %q) = %q, expected %q", tt.input, tt.length, tt.padChar, result, tt.expected)
			}
		})
	}
}

// TestIsNumeric tests IsNumeric function
func TestIsNumeric(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"empty", "", false},
		{"only digits", "12345", true},
		{"with minus", "-123", false},
		{"with decimal", "123.45", false},
		{"with comma", "1,234", false},
		{"hex digits", "123ABC", false},
		{"mixed", "123abc", false},
		{"unicode digits", "١٢٣", true}, // Arabic numerals
		{"zero", "0", true},
		{"leading zeros", "00123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNumeric(tt.input)
			if result != tt.expected {
				t.Errorf("IsNumeric(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestIsAlpha tests IsAlpha function
func TestIsAlpha(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"empty", "", false},
		{"only letters", "Hello", true},
		{"with spaces", "Hello World", false},
		{"with numbers", "Hello123", false},
		{"with punctuation", "Hello!", false},
		{"unicode letters", "café", true},
		{"mixed script", "HelloПривет", true}, // English + Russian
		{"only uppercase", "HELLO", true},
		{"only lowercase", "hello", true},
		{"emoji", "👋", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAlpha(tt.input)
			if result != tt.expected {
				t.Errorf("IsAlpha(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestIsAlphaNumeric tests IsAlphaNumeric function
func TestIsAlphaNumeric(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"empty", "", false},
		{"letters only", "Hello", true},
		{"numbers only", "123", true},
		{"mixed", "Hello123", true},
		{"with spaces", "Hello 123", false},
		{"with punctuation", "Hello!", false},
		{"unicode letters", "café123", true},
		{"unicode numbers", "١٢٣", true}, // Arabic numerals
		{"mixed unicode", "Hello١٢٣", true},
		{"underscore", "hello_world", false},
		{"dash", "hello-world", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAlphaNumeric(tt.input)
			if result != tt.expected {
				t.Errorf("IsAlphaNumeric(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestStartsWithAny tests StartsWithAny function
func TestStartsWithAny(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		prefixes []string
		expected bool
	}{
		{"starts with first", "hello world", []string{"hello", "world"}, true},
		{"starts with second", "world hello", []string{"hello", "world"}, true},
		{"doesn't start with any", "foo bar", []string{"hello", "world"}, false},
		{"empty prefixes", "hello", []string{}, false},
		{"empty string", "", []string{"test"}, false},
		{"empty string with empty prefix", "", []string{""}, true},
		{"case sensitive", "Hello", []string{"hello"}, false},
		{"unicode", "café latte", []string{"café"}, true},
		{"multiple matches", "hellohello", []string{"hello"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StartsWithAny(tt.input, tt.prefixes...)
			if result != tt.expected {
				t.Errorf("StartsWithAny(%q, %v) = %v, expected %v", tt.input, tt.prefixes, result, tt.expected)
			}
		})
	}
}

// TestEndsWithAny tests EndsWithAny function
func TestEndsWithAny(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		suffixes []string
		expected bool
	}{
		{"ends with first", "hello world", []string{"world", "hello"}, true},
		{"ends with second", "world hello", []string{"world", "hello"}, true},
		{"doesn't end with any", "foo bar", []string{"hello", "world"}, false},
		{"empty suffixes", "hello", []string{}, false},
		{"empty string", "", []string{"test"}, false},
		{"empty string with empty suffix", "", []string{""}, true},
		{"case sensitive", "Hello", []string{"hello"}, false},
		{"unicode", "latte café", []string{"café"}, true},
		{"multiple matches", "worldworld", []string{"world"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EndsWithAny(tt.input, tt.suffixes...)
			if result != tt.expected {
				t.Errorf("EndsWithAny(%q, %v) = %v, expected %v", tt.input, tt.suffixes, result, tt.expected)
			}
		})
	}
}

// TestSlice tests Slice function
func TestSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		start    int
		end      int
		expected string
	}{
		{"simple slice", "hello world", 0, 5, "hello"},
		{"negative start", "hello world", -5, -1, "worl"},
		{"negative end", "hello world", 6, -1, "worl"},
		{"both negative", "hello world", -5, -1, "worl"},
		{"out of bounds", "hello", 10, 15, ""},
		{"empty slice", "hello", 2, 2, ""},
		{"unicode slice", "café latte", 0, 4, "café"},
		{"emoji slice", "👋🌍🎉", 0, 2, "👋🌍"},
		{"full string", "hello", 0, 5, "hello"},
		{"start > end", "hello", 3, 1, ""},
		{"negative too far", "hello", -10, 3, "hel"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Slice(tt.input, tt.start, tt.end)
			if result != tt.expected {
				t.Errorf("Slice(%q, %d, %d) = %q, expected %q", tt.input, tt.start, tt.end, result, tt.expected)
			}
		})
	}
}

// TestChunk tests Chunk function
func TestChunk(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		size     int
		expected []string
	}{
		{"exact chunks", "abcdef", 2, []string{"ab", "cd", "ef"}},
		{"partial last chunk", "abcdefg", 3, []string{"abc", "def", "g"}},
		{"size larger than string", "abc", 5, []string{"abc"}},
		{"size 1", "abc", 1, []string{"a", "b", "c"}},
		{"empty string", "", 3, []string{}},
		{"zero size", "abc", 0, []string{}},
		{"negative size", "abc", -1, []string{}},
		{"unicode chunks", "cafélatte", 3, []string{"caf", "éla", "tte"}},
		{"emoji chunks", "👋🌍🎉🍕", 2, []string{"👋🌍", "🎉🍕"}},
		{"single chunk", "hello", 10, []string{"hello"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Chunk(tt.input, tt.size)
			if len(result) != len(tt.expected) {
				t.Errorf("Chunk(%q, %d) = %v, expected %v", tt.input, tt.size, result, tt.expected)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("Chunk(%q, %d)[%d] = %q, expected %q", tt.input, tt.size, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// TestLevenshteinDistance tests LevenshteinDistance function
func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		name     string
		s1       string
		s2       string
		expected int
	}{
		{"identical", "kitten", "kitten", 0},
		{"one insertion", "kitten", "kittens", 1},
		{"one deletion", "kitten", "kiten", 1},
		{"one substitution", "kitten", "sitten", 1},
		{"multiple changes", "kitten", "sitting", 3},
		{"empty first", "", "test", 4},
		{"empty second", "test", "", 4},
		{"both empty", "", "", 0},
		{"unicode", "café", "cafe", 1}, // é -> e
		{"case difference", "Hello", "hello", 1},
		{"completely different", "abc", "xyz", 3},
		{"transposition", "ab", "ba", 2}, // Levenshtein doesn't count transposition as 1
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LevenshteinDistance(tt.s1, tt.s2)
			if result != tt.expected {
				t.Errorf("LevenshteinDistance(%q, %q) = %d, expected %d", tt.s1, tt.s2, result, tt.expected)
			}
		})
	}
}

// TestSimilarity tests Similarity function
func TestSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		s1       string
		s2       string
		expected float64
	}{
		{"identical", "hello", "hello", 100.0},
		{"completely different", "abc", "xyz", 0.0},
		{"similar", "kitten", "sitting", 57.14}, // 3 edits / 7 max length
		{"empty strings", "", "", 100.0},
		{"one empty", "test", "", 0.0},
		{"case difference", "Hello", "hello", 80.0}, // 1 edit / 5 length
		{"partial match", "hello", "hell", 80.0},    // 1 deletion
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Similarity(tt.s1, tt.s2)
			// Allow small floating point differences
			epsilon := 0.01
			if result < tt.expected-epsilon || result > tt.expected+epsilon {
				t.Errorf("Similarity(%q, %q) = %.2f, expected %.2f", tt.s1, tt.s2, result, tt.expected)
			}
		})
	}
}

// TestUnique tests Unique function
func TestUnique(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{"all unique", []string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{"duplicates", []string{"a", "b", "a", "c", "b"}, []string{"a", "b", "c"}},
		{"empty slice", []string{}, []string{}},
		{"all same", []string{"a", "a", "a"}, []string{"a"}},
		{"preserves order", []string{"b", "a", "c", "a", "b"}, []string{"b", "a", "c"}},
		{"case sensitive", []string{"A", "a", "B", "b"}, []string{"A", "a", "B", "b"}},
		{"unicode", []string{"café", "cafe", "café"}, []string{"café", "cafe"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Unique(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Unique(%v) = %v, expected %v", tt.input, result, tt.expected)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("Unique(%v)[%d] = %q, expected %q", tt.input, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// TestMask tests Mask function
func TestMask(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		visibleChars int
		maskChar     rune
		expected     string
	}{
		{"mask all", "password", 0, '*', "********"},
		{"show first 2", "password", 2, '*', "pa******"},
		{"show first 4", "1234567890", 4, '#', "1234######"},
		{"no masking needed", "short", 5, '*', "short"},
		{"visible > length", "test", 10, '*', "test"},
		{"empty string", "", 3, '*', ""},
		{"mask char different", "password", 3, 'X', "pasXXXXX"},
		{"unicode string", "café latte", 4, '*', "café******"},
		{"emoji masking", "👋🌍🎉", 1, '?', "👋??"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Mask(tt.input, tt.visibleChars, tt.maskChar)
			if result != tt.expected {
				t.Errorf("Mask(%q, %d, %q) = %q, expected %q", tt.input, tt.visibleChars, tt.maskChar, result, tt.expected)
			}
		})
	}
}

// TestMaskEmail tests MaskEmail function
func TestMaskEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal email", "john.doe@example.com", "j********e@example.com"},
		{"short local", "ab@example.com", "ab@example.com"},
		{"3 char local", "abc@example.com", "a*c@example.com"},
		{"single char", "a@example.com", "a@example.com"},
		{"invalid email", "notanemail", "notanemail"},
		{"no @ symbol", "example.com", "example.com"},
		{"multiple @", "a@b@c.com", "a@b@c.com"},
		{"with plus", "john+tag@example.com", "j********g@example.com"},
		{"with numbers", "user123@example.com", "u******3@example.com"},
		{"empty local", "@example.com", "@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskEmail(tt.input)
			if result != tt.expected {
				t.Errorf("MaskEmail(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestGenerateRandomString tests GenerateRandomString function
func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"zero length", 0},
		{"length 1", 1},
		{"length 10", 10},
		{"length 100", 100},
		{"length 1000", 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateRandomString(tt.length)

			// Check length
			if len(result) != tt.length {
				t.Errorf("GenerateRandomString(%d) length = %d, expected %d", tt.length, len(result), tt.length)
			}

			// For non-zero length, check that strings are different
			if tt.length > 0 {
				result2 := GenerateRandomString(tt.length)
				if result == result2 {
					t.Errorf("GenerateRandomString(%d) produced identical strings: %q", tt.length, result)
				}

				// Check that it contains only allowed characters
				allowedChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
				for _, ch := range result {
					if !strings.ContainsRune(allowedChars, ch) {
						t.Errorf("GenerateRandomString(%d) contains invalid character: %q", tt.length, ch)
					}
				}
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkReverse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Reverse("hello world this is a test string")
	}
}

func BenchmarkIsPalindrome(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsPalindrome("A man a plan a canal Panama")
	}
}

func BenchmarkToCamelCase(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ToCamelCase("hello_world_test_string_example")
	}
}

func BenchmarkLevenshteinDistance(b *testing.B) {
	for i := 0; i < b.N; i++ {
		LevenshteinDistance("kitten", "sitting")
	}
}

// Example tests (documentation examples)
func ExampleReverse() {
	fmt.Println(Reverse("hello"))
	// Output: olleh
}

func ExampleIsPalindrome() {
	fmt.Println(IsPalindrome("racecar"))
	fmt.Println(IsPalindrome("hello"))
	// Output:
	// true
	// false
}

func ExampleToCamelCase() {
	fmt.Println(ToCamelCase("hello_world"))
	fmt.Println(ToCamelCase("hello-world"))
	fmt.Println(ToCamelCase("Hello World"))
	// Output:
	// helloWorld
	// helloWorld
	// helloWorld
}

func ExampleMaskEmail() {
	fmt.Println(MaskEmail("john.doe@example.com"))
	// Output: j********e@example.com
}
