package numbersutils

import (
	"math"
	"math/big"
	"sort"
	"testing"
)

// Test Min/Max functions
func TestMin(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{1, 2, 1},
		{5, 3, 3},
		{-1, 1, -1},
		{0, 0, 0},
		{-5, -10, -10},
	}

	for _, tt := range tests {
		result := Min(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("Min(%d, %d) = %d; expected %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{1, 2, 2},
		{5, 3, 5},
		{-1, 1, 1},
		{0, 0, 0},
		{-5, -10, -5},
	}

	for _, tt := range tests {
		result := Max(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("Max(%d, %d) = %d; expected %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestMinInt64(t *testing.T) {
	tests := []struct {
		a, b, expected int64
	}{
		{1, 2, 1},
		{5, 3, 3},
		{-1, 1, -1},
		{0, 0, 0},
		{-5, -10, -10},
	}

	for _, tt := range tests {
		result := MinInt64(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("MinInt64(%d, %d) = %d; expected %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestMaxInt64(t *testing.T) {
	tests := []struct {
		a, b, expected int64
	}{
		{1, 2, 2},
		{5, 3, 5},
		{-1, 1, 1},
		{0, 0, 0},
		{-5, -10, -5},
	}

	for _, tt := range tests {
		result := MaxInt64(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("MaxInt64(%d, %d) = %d; expected %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

// Test Abs functions
func TestAbs(t *testing.T) {
	tests := []struct {
		x, expected int
	}{
		{5, 5},
		{-5, 5},
		{0, 0},
		{-100, 100},
		{100, 100},
	}

	for _, tt := range tests {
		result := Abs(tt.x)
		if result != tt.expected {
			t.Errorf("Abs(%d) = %d; expected %d", tt.x, result, tt.expected)
		}
	}
}

func TestAbsFloat64(t *testing.T) {
	tests := []struct {
		x, expected float64
	}{
		{5.5, 5.5},
		{-5.5, 5.5},
		{0.0, 0.0},
		{-100.25, 100.25},
		{100.25, 100.25},
	}

	for _, tt := range tests {
		result := AbsFloat64(tt.x)
		if result != tt.expected {
			t.Errorf("AbsFloat64(%f) = %f; expected %f", tt.x, result, tt.expected)
		}
	}
}

// Test Sum functions
func TestSum(t *testing.T) {
	tests := []struct {
		numbers  []int
		expected int
	}{
		{[]int{1, 2, 3, 4, 5}, 15},
		{[]int{-1, -2, -3}, -6},
		{[]int{}, 0},
		{[]int{10}, 10},
		{[]int{0, 0, 0}, 0},
	}

	for _, tt := range tests {
		result := Sum(tt.numbers)
		if result != tt.expected {
			t.Errorf("Sum(%v) = %d; expected %d", tt.numbers, result, tt.expected)
		}
	}
}

func TestSumFloat64(t *testing.T) {
	tests := []struct {
		numbers  []float64
		expected float64
	}{
		{[]float64{1.5, 2.5, 3.5}, 7.5},
		{[]float64{-1.5, -2.5}, -4.0},
		{[]float64{}, 0},
		{[]float64{10.1}, 10.1},
		{[]float64{0.0, 0.0, 0.0}, 0.0},
	}

	for _, tt := range tests {
		result := SumFloat64(tt.numbers)
		if math.Abs(result-tt.expected) > 1e-10 {
			t.Errorf("SumFloat64(%v) = %f; expected %f", tt.numbers, result, tt.expected)
		}
	}
}

// Test Average functions
func TestAverage(t *testing.T) {
	tests := []struct {
		numbers  []int
		expected float64
	}{
		{[]int{1, 2, 3, 4, 5}, 3.0},
		{[]int{-1, -2, -3}, -2.0},
		{[]int{}, 0.0},
		{[]int{10}, 10.0},
		{[]int{0, 0, 0}, 0.0},
	}

	for _, tt := range tests {
		result := Average(tt.numbers)
		if math.Abs(result-tt.expected) > 1e-10 {
			t.Errorf("Average(%v) = %f; expected %f", tt.numbers, result, tt.expected)
		}
	}
}

// Test Median functions
func TestMedian(t *testing.T) {
	tests := []struct {
		numbers  []int
		expected float64
	}{
		{[]int{1, 2, 3, 4, 5}, 3.0},
		{[]int{1, 2, 3, 4, 5, 6}, 3.5},
		{[]int{10}, 10.0},
		{[]int{}, 0.0},
		{[]int{5, 1, 3, 2, 4}, 3.0}, // Unsorted
	}

	for _, tt := range tests {
		result := Median(tt.numbers)
		if math.Abs(result-tt.expected) > 1e-10 {
			t.Errorf("Median(%v) = %f; expected %f", tt.numbers, result, tt.expected)
		}
	}
}

func TestMedianFloat64(t *testing.T) {
	tests := []struct {
		numbers  []float64
		expected float64
	}{
		{[]float64{1.5, 2.5, 3.5, 4.5, 5.5}, 3.5},
		{[]float64{1.1, 2.2, 3.3, 4.4, 5.5, 6.6}, 3.85},
		{[]float64{10.5}, 10.5},
		{[]float64{}, 0.0},
	}

	for _, tt := range tests {
		result := MedianFloat64(tt.numbers)
		if math.Abs(result-tt.expected) > 1e-10 {
			t.Errorf("MedianFloat64(%v) = %f; expected %f", tt.numbers, result, tt.expected)
		}
	}
}

// Test Mode function
func TestMode(t *testing.T) {
	tests := []struct {
		numbers  []int
		expected []int
	}{
		{[]int{1, 2, 2, 3, 4}, []int{2}},
		{[]int{1, 1, 2, 2, 3}, []int{1, 2}},
		{[]int{1}, []int{1}},
		{[]int{}, []int{}},
		{[]int{5, 5, 5, 5}, []int{5}},
	}

	for _, tt := range tests {
		result := Mode(tt.numbers)
		if len(result) != len(tt.expected) {
			t.Errorf("Mode(%v) = %v; expected %v", tt.numbers, result, tt.expected)
			continue
		}
		for i := range result {
			if result[i] != tt.expected[i] {
				t.Errorf("Mode(%v) = %v; expected %v", tt.numbers, result, tt.expected)
				break
			}
		}
	}
}

// Test Standard Deviation functions
func TestStandardDeviation(t *testing.T) {
	tests := []struct {
		numbers  []int
		expected float64
	}{
		{[]int{1, 2, 3, 4, 5}, 1.5811388300841898},
		{[]int{10, 10, 10, 10}, 0.0},
		{[]int{1}, 0.0},
	}

	for _, tt := range tests {
		result := StandardDeviation(tt.numbers)
		if math.Abs(result-tt.expected) > 1e-10 {
			t.Errorf("StandardDeviation(%v) = %f; expected %f", tt.numbers, result, tt.expected)
		}
	}
}

// Test Prime functions
func TestIsPrime(t *testing.T) {
	tests := []struct {
		n        int
		expected bool
	}{
		{2, true},
		{3, true},
		{5, true},
		{7, true},
		{11, true},
		{4, false},
		{6, false},
		{9, false},
		{1, false},
		{0, false},
		{-5, false},
	}

	for _, tt := range tests {
		result := IsPrime(tt.n)
		if result != tt.expected {
			t.Errorf("IsPrime(%d) = %v; expected %v", tt.n, result, tt.expected)
		}
	}
}

// Test Factorial functions
func TestFactorial(t *testing.T) {
	tests := []struct {
		n        int
		expected int
	}{
		{0, 1},
		{1, 1},
		{5, 120},
		{7, 5040},
		{10, 3628800},
	}

	for _, tt := range tests {
		result := Factorial(tt.n)
		if result != tt.expected {
			t.Errorf("Factorial(%d) = %d; expected %d", tt.n, result, tt.expected)
		}
	}
}

func TestFactorialBig(t *testing.T) {
	tests := []struct {
		n        int64
		expected *big.Int
	}{
		{0, big.NewInt(1)},
		{1, big.NewInt(1)},
		{5, big.NewInt(120)},
		{10, big.NewInt(3628800)},
		{20, big.NewInt(2432902008176640000)},
	}

	for _, tt := range tests {
		result := FactorialBig(tt.n)
		if result.Cmp(tt.expected) != 0 {
			t.Errorf("FactorialBig(%d) = %s; expected %s", tt.n, result.String(), tt.expected.String())
		}
	}
}

// Test Fibonacci functions
func TestFibonacci(t *testing.T) {
	tests := []struct {
		n        int
		expected int
	}{
		{0, 0},
		{1, 1},
		{2, 1},
		{3, 2},
		{5, 5},
		{10, 55},
		{20, 6765},
	}

	for _, tt := range tests {
		result := Fibonacci(tt.n)
		if result != tt.expected {
			t.Errorf("Fibonacci(%d) = %d; expected %d", tt.n, result, tt.expected)
		}
	}
}

func TestFibonacciSequence(t *testing.T) {
	tests := []struct {
		n        int
		expected []int
	}{
		{0, []int{}},
		{1, []int{0}},
		{2, []int{0, 1}},
		{5, []int{0, 1, 1, 2, 3}},
		{10, []int{0, 1, 1, 2, 3, 5, 8, 13, 21, 34}},
	}

	for _, tt := range tests {
		result := FibonacciSequence(tt.n)
		if len(result) != len(tt.expected) {
			t.Errorf("FibonacciSequence(%d) length = %d; expected %d", tt.n, len(result), len(tt.expected))
			continue
		}
		for i := range result {
			if result[i] != tt.expected[i] {
				t.Errorf("FibonacciSequence(%d)[%d] = %d; expected %d", tt.n, i, result[i], tt.expected[i])
				break
			}
		}
	}
}

// Test Round functions
func TestRound(t *testing.T) {
	tests := []struct {
		value    float64
		decimals int
		expected float64
	}{
		{3.14159, 2, 3.14},
		{3.14159, 3, 3.142},
		{3.14159, 0, 3.0},
		{2.5, 0, 3.0},
		{-2.5, 0, -3.0},
		{0.0, 5, 0.0},
		{123.456789, 4, 123.4568},
	}

	for _, tt := range tests {
		result := Round(tt.value, tt.decimals)
		if math.Abs(result-tt.expected) > 1e-10 {
			t.Errorf("Round(%f, %d) = %f; expected %f", tt.value, tt.decimals, result, tt.expected)
		}
	}
}

func TestRoundUp(t *testing.T) {
	tests := []struct {
		value    float64
		decimals int
		expected float64
	}{
		{3.14159, 2, 3.15},
		{3.14159, 3, 3.142},
		{2.1, 0, 3.0},
		{-2.1, 0, -2.0},
		{0.0, 5, 0.0},
	}

	for _, tt := range tests {
		result := RoundUp(tt.value, tt.decimals)
		if math.Abs(result-tt.expected) > 1e-10 {
			t.Errorf("RoundUp(%f, %d) = %f; expected %f", tt.value, tt.decimals, result, tt.expected)
		}
	}
}

func TestRoundDown(t *testing.T) {
	tests := []struct {
		value    float64
		decimals int
		expected float64
	}{
		{3.14159, 2, 3.14},
		{3.14159, 3, 3.141},
		{2.9, 0, 2.0},
		{-2.9, 0, -3.0},
		{0.0, 5, 0.0},
	}

	for _, tt := range tests {
		result := RoundDown(tt.value, tt.decimals)
		if math.Abs(result-tt.expected) > 1e-10 {
			t.Errorf("RoundDown(%f, %d) = %f; expected %f", tt.value, tt.decimals, result, tt.expected)
		}
	}
}

// Test Clamp functions
func TestClamp(t *testing.T) {
	tests := []struct {
		value, min, max, expected int
	}{
		{5, 1, 10, 5},
		{0, 1, 10, 1},
		{15, 1, 10, 10},
		{-5, -10, 10, -5},
		{-15, -10, 10, -10},
		{20, -10, 10, 10},
	}

	for _, tt := range tests {
		result := Clamp(tt.value, tt.min, tt.max)
		if result != tt.expected {
			t.Errorf("Clamp(%d, %d, %d) = %d; expected %d", tt.value, tt.min, tt.max, result, tt.expected)
		}
	}
}

// Test GCD and LCM functions
func TestGCD(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{48, 18, 6},
		{17, 5, 1},
		{0, 5, 5},
		{5, 0, 5},
		{0, 0, 0},
		{-48, 18, 6},
	}

	for _, tt := range tests {
		result := GCD(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("GCD(%d, %d) = %d; expected %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestLCM(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{12, 18, 36},
		{5, 7, 35},
		{0, 5, 0},
		{5, 0, 0},
		{0, 0, 0},
	}

	for _, tt := range tests {
		result := LCM(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("LCM(%d, %d) = %d; expected %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

// Test Percent functions
func TestPercentChange(t *testing.T) {
	tests := []struct {
		oldValue, newValue, expected float64
	}{
		{100, 120, 20.0},
		{100, 80, -20.0},
		{50, 100, 100.0},
		{0, 100, 100.0},
		{100, 0, -100.0},
		{0, 0, 0.0},
	}

	for _, tt := range tests {
		result := PercentChange(tt.oldValue, tt.newValue)
		if math.Abs(result-tt.expected) > 1e-10 {
			t.Errorf("PercentChange(%f, %f) = %f; expected %f", tt.oldValue, tt.newValue, result, tt.expected)
		}
	}
}

func TestPercentOf(t *testing.T) {
	tests := []struct {
		a, b, expected float64
	}{
		{50, 100, 50.0},
		{25, 200, 12.5},
		{0, 100, 0.0},
		{100, 0, 0.0},
	}

	for _, tt := range tests {
		result := PercentOf(tt.a, tt.b)
		if math.Abs(result-tt.expected) > 1e-10 {
			t.Errorf("PercentOf(%f, %f) = %f; expected %f", tt.a, tt.b, result, tt.expected)
		}
	}
}

// Test Number properties
func TestIsPerfectSquare(t *testing.T) {
	tests := []struct {
		n        int
		expected bool
	}{
		{1, true},
		{4, true},
		{9, true},
		{16, true},
		{25, true},
		{2, false},
		{3, false},
		{5, false},
		{-1, false},
		{-4, false},
	}

	for _, tt := range tests {
		result := IsPerfectSquare(tt.n)
		if result != tt.expected {
			t.Errorf("IsPerfectSquare(%d) = %v; expected %v", tt.n, result, tt.expected)
		}
	}
}

func TestIsPowerOfTwo(t *testing.T) {
	tests := []struct {
		n        int
		expected bool
	}{
		{1, true},
		{2, true},
		{4, true},
		{8, true},
		{16, true},
		{3, false},
		{5, false},
		{6, false},
		{0, false},
		{-2, false},
	}

	for _, tt := range tests {
		result := IsPowerOfTwo(tt.n)
		if result != tt.expected {
			t.Errorf("IsPowerOfTwo(%d) = %v; expected %v", tt.n, result, tt.expected)
		}
	}
}

func TestIsArmstrong(t *testing.T) {
	tests := []struct {
		num      int
		expected bool
	}{
		{153, true},   // 1^3 + 5^3 + 3^3 = 153
		{370, true},   // 3^3 + 7^3 + 0^3 = 370
		{371, true},   // 3^3 + 7^3 + 1^3 = 371
		{407, true},   // 4^3 + 0^3 + 7^3 = 407
		{1634, true},  // 1^4 + 6^4 + 3^4 + 4^4 = 1634
		{123, false},  // 1^3 + 2^3 + 3^3 = 36 != 123
		{0, true},     // 0^1 = 0
		{1, true},     // 1^1 = 1
		{-153, false}, // Negative numbers are not Armstrong
	}

	for _, tt := range tests {
		result := IsArmstrong(tt.num)
		if result != tt.expected {
			t.Errorf("IsArmstrong(%d) = %v; expected %v", tt.num, result, tt.expected)
		}
	}
}

func TestIsPalindromeNumber(t *testing.T) {
	tests := []struct {
		num      int
		expected bool
	}{
		{121, true},
		{12321, true},
		{123321, true},
		{123, false},
		{-121, false},
		{0, true},
		{9, true},
		{10, false},
	}

	for _, tt := range tests {
		result := IsPalindromeNumber(tt.num)
		if result != tt.expected {
			t.Errorf("IsPalindromeNumber(%d) = %v; expected %v", tt.num, result, tt.expected)
		}
	}
}

// Test Digit functions
func TestDigits(t *testing.T) {
	tests := []struct {
		num      int
		expected []int
	}{
		{12345, []int{1, 2, 3, 4, 5}},
		{0, []int{0}},
		{987654321, []int{9, 8, 7, 6, 5, 4, 3, 2, 1}},
		{-123, []int{1, 2, 3}},
		{7, []int{7}},
	}

	for _, tt := range tests {
		result := Digits(tt.num)
		if len(result) != len(tt.expected) {
			t.Errorf("Digits(%d) length = %d; expected %d", tt.num, len(result), len(tt.expected))
			continue
		}
		for i := range result {
			if result[i] != tt.expected[i] {
				t.Errorf("Digits(%d)[%d] = %d; expected %d", tt.num, i, result[i], tt.expected[i])
				break
			}
		}
	}
}

func TestDigitSum(t *testing.T) {
	tests := []struct {
		num      int
		expected int
	}{
		{12345, 15},
		{0, 0},
		{987654321, 45},
		{-123, 6},
		{999, 27},
	}

	for _, tt := range tests {
		result := DigitSum(tt.num)
		if result != tt.expected {
			t.Errorf("DigitSum(%d) = %d; expected %d", tt.num, result, tt.expected)
		}
	}
}

// Test Base conversion functions
func TestToBinary(t *testing.T) {
	tests := []struct {
		n        int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{2, "10"},
		{5, "101"},
		{10, "1010"},
		{255, "11111111"},
		{-5, "-101"},
	}

	for _, tt := range tests {
		result := ToBinary(tt.n)
		if result != tt.expected {
			t.Errorf("ToBinary(%d) = %s; expected %s", tt.n, result, tt.expected)
		}
	}
}

func TestToHex(t *testing.T) {
	tests := []struct {
		n        int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{10, "a"},
		{15, "f"},
		{16, "10"},
		{255, "ff"},
		{-255, "-ff"},
	}

	for _, tt := range tests {
		result := ToHex(tt.n)
		if result != tt.expected {
			t.Errorf("ToHex(%d) = %s; expected %s", tt.n, result, tt.expected)
		}
	}
}

func TestFromBinary(t *testing.T) {
	tests := []struct {
		bin         string
		expected    int
		expectError bool
	}{
		{"0", 0, false},
		{"1", 1, false},
		{"1010", 10, false},
		{"11111111", 255, false},
		{"-101", -5, false},
		{"abc", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		result, err := FromBinary(tt.bin)
		if tt.expectError && err == nil {
			t.Errorf("FromBinary(%s) expected error but got none", tt.bin)
		}
		if !tt.expectError && err != nil {
			t.Errorf("FromBinary(%s) unexpected error: %v", tt.bin, err)
		}
		if !tt.expectError && result != tt.expected {
			t.Errorf("FromBinary(%s) = %d; expected %d", tt.bin, result, tt.expected)
		}
	}
}

// Test Validation functions
func TestIsValidCreditCard(t *testing.T) {
	tests := []struct {
		number   string
		expected bool
	}{
		{"4111111111111111", true},  // Visa test number
		{"4111111111111", true},     // 13-digit Visa
		{"4012888888881881", true},  // Visa test number
		{"378282246310005", true},   // American Express
		{"6011111111111117", true},  // Discover
		{"5555555555554444", true},  // MasterCard
		{"1234567890123456", false}, // Invalid
		{"4111111111111112", false}, // Invalid checksum
		{"", false},                 // Empty
		{"abc", false},              // Non-digits
	}

	for _, tt := range tests {
		result := IsValidCreditCard(tt.number)
		if result != tt.expected {
			t.Errorf("IsValidCreditCard(%s) = %v; expected %v", tt.number, result, tt.expected)
		}
	}
}

func TestIsValidISBN(t *testing.T) {
	tests := []struct {
		isbn     string
		expected bool
	}{
		{"0-306-40615-2", true},     // Valid ISBN-10
		{"0306406152", true},        // Valid ISBN-10 without hyphens
		{"978-3-16-148410-0", true}, // Valid ISBN-13
		{"9783161484100", true},     // Valid ISBN-13 without hyphens
		{"1234567890", false},       // Invalid ISBN-10
		{"1234567890123", false},    // Invalid ISBN-13
		{"", false},                 // Empty
		{"abc", false},              // Invalid format
	}

	for _, tt := range tests {
		result := IsValidISBN(tt.isbn)
		if result != tt.expected {
			t.Errorf("IsValidISBN(%s) = %v; expected %v", tt.isbn, result, tt.expected)
		}
	}
}

func TestIsValidPhoneNumber(t *testing.T) {
	tests := []struct {
		phone    string
		expected bool
	}{
		{"1234567", true},
		{"123456789012345", true},
		{"123456", false},           // Too short
		{"1234567890123456", false}, // Too long
		{"(123) 456-7890", true},    // With formatting
		{"123-456-7890", true},      // With hyphens
		{"", false},                 // Empty
		{"abc", false},              // Non-digits
	}

	for _, tt := range tests {
		result := IsValidPhoneNumber(tt.phone)
		if result != tt.expected {
			t.Errorf("IsValidPhoneNumber(%s) = %v; expected %v", tt.phone, result, tt.expected)
		}
	}
}

func TestIsValidSSN(t *testing.T) {
	tests := []struct {
		ssn      string
		expected bool
	}{
		{"123-45-6789", true},
		{"000-45-6789", false}, // Invalid first part
		{"666-45-6789", true},  // Valid (666 is allowed)
		{"123-00-6789", false}, // Invalid middle part
		{"123-45-0000", false}, // Invalid last part
		{"000-00-0000", false}, // All zeros
		{"123456789", false},   // Missing hyphens
		{"abc-de-fghi", false}, // Non-digits
		{"", false},            // Empty
	}

	for _, tt := range tests {
		result := IsValidSSN(tt.ssn)
		if result != tt.expected {
			t.Errorf("IsValidSSN(%s) = %v; expected %v", tt.ssn, result, tt.expected)
		}
	}
}

// Test Array operations
func TestUnique(t *testing.T) {
	tests := []struct {
		numbers  []int
		expected []int
	}{
		{[]int{1, 2, 3, 2, 1}, []int{1, 2, 3}},
		{[]int{5, 5, 5, 5}, []int{5}},
		{[]int{}, []int{}},
		{[]int{1, 2, 3}, []int{1, 2, 3}},
		{[]int{3, 1, 2, 1, 3}, []int{3, 1, 2}},
	}

	for _, tt := range tests {
		result := Unique(tt.numbers)
		if len(result) != len(tt.expected) {
			t.Errorf("Unique(%v) length = %d; expected %d", tt.numbers, len(result), len(tt.expected))
			continue
		}

		// Sort both slices for comparison since order doesn't matter
		sortInts(result)
		sortInts(tt.expected)

		for i := range result {
			if result[i] != tt.expected[i] {
				t.Errorf("Unique(%v)[%d] = %d; expected %d", tt.numbers, i, result[i], tt.expected[i])
				break
			}
		}
	}
}

func TestIntersection(t *testing.T) {
	tests := []struct {
		a, b     []int
		expected []int
	}{
		{[]int{1, 2, 3}, []int{2, 3, 4}, []int{2, 3}},
		{[]int{1, 2, 3}, []int{4, 5, 6}, []int{}},
		{[]int{}, []int{1, 2, 3}, []int{}},
		{[]int{1, 2, 2, 3}, []int{2, 2, 3, 4}, []int{2, 3}},
		{[]int{1, 2, 3}, []int{1, 2, 3}, []int{1, 2, 3}},
	}

	for _, tt := range tests {
		result := Intersection(tt.a, tt.b)
		if len(result) != len(tt.expected) {
			t.Errorf("Intersection(%v, %v) length = %d; expected %d", tt.a, tt.b, len(result), len(tt.expected))
			continue
		}

		sortInts(result)
		sortInts(tt.expected)

		for i := range result {
			if result[i] != tt.expected[i] {
				t.Errorf("Intersection(%v, %v)[%d] = %d; expected %d", tt.a, tt.b, i, result[i], tt.expected[i])
				break
			}
		}
	}
}

func TestUnion(t *testing.T) {
	tests := []struct {
		a, b     []int
		expected []int
	}{
		{[]int{1, 2, 3}, []int{2, 3, 4}, []int{1, 2, 3, 4}},
		{[]int{1, 2}, []int{3, 4}, []int{1, 2, 3, 4}},
		{[]int{}, []int{1, 2, 3}, []int{1, 2, 3}},
		{[]int{1, 2, 2}, []int{2, 2, 3}, []int{1, 2, 3}},
	}

	for _, tt := range tests {
		result := Union(tt.a, tt.b)
		if len(result) != len(tt.expected) {
			t.Errorf("Union(%v, %v) length = %d; expected %d", tt.a, tt.b, len(result), len(tt.expected))
			continue
		}

		sortInts(result)
		sortInts(tt.expected)

		for i := range result {
			if result[i] != tt.expected[i] {
				t.Errorf("Union(%v, %v)[%d] = %d; expected %d", tt.a, tt.b, i, result[i], tt.expected[i])
				break
			}
		}
	}
}

// Test Range functions
func TestRange(t *testing.T) {
	tests := []struct {
		start, end, step int
		expected         []int
	}{
		{0, 5, 1, []int{0, 1, 2, 3, 4}},
		{5, 0, -1, []int{5, 4, 3, 2, 1}},
		{0, 10, 2, []int{0, 2, 4, 6, 8}},
		{10, 0, -2, []int{10, 8, 6, 4, 2}},
		{0, 0, 1, []int{}},
		{5, 5, 1, []int{}},
	}

	for _, tt := range tests {
		result := Range(tt.start, tt.end, tt.step)
		if len(result) != len(tt.expected) {
			t.Errorf("Range(%d, %d, %d) length = %d; expected %d", tt.start, tt.end, tt.step, len(result), len(tt.expected))
			continue
		}
		for i := range result {
			if result[i] != tt.expected[i] {
				t.Errorf("Range(%d, %d, %d)[%d] = %d; expected %d", tt.start, tt.end, tt.step, i, result[i], tt.expected[i])
				break
			}
		}
	}
}

func TestLinspace(t *testing.T) {
	tests := []struct {
		start, end float64
		n          int
		expected   []float64
	}{
		{0.0, 1.0, 5, []float64{0.0, 0.25, 0.5, 0.75, 1.0}},
		{1.0, 0.0, 5, []float64{1.0, 0.75, 0.5, 0.25, 0.0}},
		{0.0, 10.0, 3, []float64{0.0, 5.0, 10.0}},
		{5.0, 5.0, 1, []float64{5.0}},
		{0.0, 0.0, 0, []float64{}},
	}

	for _, tt := range tests {
		result := Linspace(tt.start, tt.end, tt.n)
		if len(result) != len(tt.expected) {
			t.Errorf("Linspace(%f, %f, %d) length = %d; expected %d", tt.start, tt.end, tt.n, len(result), len(tt.expected))
			continue
		}
		for i := range result {
			if math.Abs(result[i]-tt.expected[i]) > 1e-10 {
				t.Errorf("Linspace(%f, %f, %d)[%d] = %f; expected %f", tt.start, tt.end, tt.n, i, result[i], tt.expected[i])
				break
			}
		}
	}
}

// Test Normalization functions
func TestNormalize(t *testing.T) {
	tests := []struct {
		numbers  []float64
		expected []float64
	}{
		{[]float64{1.0, 2.0, 3.0, 4.0, 5.0}, []float64{0.0, 0.25, 0.5, 0.75, 1.0}},
		{[]float64{5.0, 5.0, 5.0}, []float64{0.0, 0.0, 0.0}},
		{[]float64{10.0}, []float64{0.0}},
		{[]float64{}, []float64{}},
		{[]float64{-5.0, 0.0, 5.0}, []float64{0.0, 0.5, 1.0}},
	}

	for _, tt := range tests {
		result := Normalize(tt.numbers)
		if len(result) != len(tt.expected) {
			t.Errorf("Normalize(%v) length = %d; expected %d", tt.numbers, len(result), len(tt.expected))
			continue
		}
		for i := range result {
			if math.Abs(result[i]-tt.expected[i]) > 1e-10 {
				t.Errorf("Normalize(%v)[%d] = %f; expected %f", tt.numbers, i, result[i], tt.expected[i])
				break
			}
		}
	}
}

func TestStandardize(t *testing.T) {
	numbers := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	result := Standardize(numbers)

	// Mean should be approximately 0
	mean := AverageFloat64(result)
	if math.Abs(mean) > 1e-10 {
		t.Errorf("Standardize() mean = %f; expected ~0", mean)
	}

	// Standard deviation should be approximately 1
	std := StandardDeviationFloat64(result)
	if math.Abs(std-1.0) > 1e-10 {
		t.Errorf("Standardize() std = %f; expected ~1", std)
	}
}

// Test Quantile functions
func TestQuantile(t *testing.T) {
	tests := []struct {
		numbers  []float64
		q        float64
		expected float64
	}{
		{[]float64{1, 2, 3, 4, 5}, 0.5, 3.0},
		{[]float64{1, 2, 3, 4, 5}, 0.25, 1.5},
		{[]float64{1, 2, 3, 4, 5}, 0.75, 4.5},
		{[]float64{1, 2, 3, 4, 5}, 0.0, 1.0},
		{[]float64{1, 2, 3, 4, 5}, 1.0, 5.0},
		{[]float64{10.0}, 0.5, 10.0},
	}

	for _, tt := range tests {
		result := Quantile(tt.numbers, tt.q)
		if math.Abs(result-tt.expected) > 1e-10 {
			t.Errorf("Quantile(%v, %f) = %f; expected %f", tt.numbers, tt.q, result, tt.expected)
		}
	}
}

func TestQuartiles(t *testing.T) {
	numbers := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9}
	q1, q2, q3 := Quartiles(numbers)

	if math.Abs(q1-2.5) > 1e-10 {
		t.Errorf("Quartiles() Q1 = %f; expected 2.5", q1)
	}
	if math.Abs(q2-5.0) > 1e-10 {
		t.Errorf("Quartiles() Q2 = %f; expected 5.0", q2)
	}
	if math.Abs(q3-7.5) > 1e-10 {
		t.Errorf("Quartiles() Q3 = %f; expected 7.5", q3)
	}
}

// Test Math functions
func TestDotProduct(t *testing.T) {
	tests := []struct {
		a, b     []float64
		expected float64
	}{
		{[]float64{1, 2, 3}, []float64{4, 5, 6}, 32.0},
		{[]float64{0, 0, 0}, []float64{1, 2, 3}, 0.0},
		{[]float64{-1, -2, -3}, []float64{1, 2, 3}, -14.0},
		{[]float64{}, []float64{}, 0.0},
		{[]float64{2.5, 3.5}, []float64{1.5, 2.5}, 12.5},
	}

	for _, tt := range tests {
		result := DotProduct(tt.a, tt.b)
		if math.Abs(result-tt.expected) > 1e-10 {
			t.Errorf("DotProduct(%v, %v) = %f; expected %f", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestEuclideanDistance(t *testing.T) {
	tests := []struct {
		a, b     []float64
		expected float64
	}{
		{[]float64{0, 0}, []float64{3, 4}, 5.0},
		{[]float64{1, 2, 3}, []float64{4, 5, 6}, 5.196152422706632},
		{[]float64{0, 0}, []float64{0, 0}, 0.0},
		{[]float64{-1, -1}, []float64{1, 1}, 2.8284271247461903},
	}

	for _, tt := range tests {
		result := EuclideanDistance(tt.a, tt.b)
		if math.Abs(result-tt.expected) > 1e-10 {
			t.Errorf("EuclideanDistance(%v, %v) = %f; expected %f", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		a, b     []float64
		expected float64
	}{
		{[]float64{1, 0}, []float64{1, 0}, 1.0},
		{[]float64{1, 0}, []float64{0, 1}, 0.0},
		{[]float64{1, 1}, []float64{1, 1}, 1.0},
		{[]float64{1, 2, 3}, []float64{4, 5, 6}, 0.9746318461970762},
		{[]float64{0, 0}, []float64{0, 0}, 0.0},
	}

	for _, tt := range tests {
		result := CosineSimilarity(tt.a, tt.b)
		if math.Abs(result-tt.expected) > 1e-10 {
			t.Errorf("CosineSimilarity(%v, %v) = %f; expected %f", tt.a, tt.b, result, tt.expected)
		}
	}
}

// Test Activation functions
func TestSigmoid(t *testing.T) {
	tests := []struct {
		x        float64
		expected float64
	}{
		{0.0, 0.5},
		{1.0, 0.7310585786300049},
		{-1.0, 0.2689414213699951},
		{10.0, 0.9999546021312976},
		{-10.0, 0.000045397868702},
	}

	for _, tt := range tests {
		result := Sigmoid(tt.x)
		if math.Abs(result-tt.expected) > 1e-10 {
			t.Errorf("Sigmoid(%f) = %f; expected %f", tt.x, result, tt.expected)
		}
	}
}

func TestSoftmax(t *testing.T) {
	tests := []struct {
		numbers  []float64
		expected []float64
	}{
		{[]float64{1.0, 2.0, 3.0}, []float64{0.09003057317038046, 0.24472847105479767, 0.6652409557748219}},
		{[]float64{0.0, 0.0, 0.0}, []float64{0.3333333333333333, 0.3333333333333333, 0.3333333333333333}},
		{[]float64{100.0, 100.0, 100.0}, []float64{0.3333333333333333, 0.3333333333333333, 0.3333333333333333}},
		{[]float64{-1.0, 0.0, 1.0}, []float64{0.09003057317038046, 0.24472847105479767, 0.6652409557748219}},
	}

	for _, tt := range tests {
		result := Softmax(tt.numbers)
		if len(result) != len(tt.expected) {
			t.Errorf("Softmax(%v) length = %d; expected %d", tt.numbers, len(result), len(tt.expected))
			continue
		}
		for i := range result {
			if math.Abs(result[i]-tt.expected[i]) > 1e-10 {
				t.Errorf("Softmax(%v)[%d] = %f; expected %f", tt.numbers, i, result[i], tt.expected[i])
				break
			}
		}
	}
}

func TestReLU(t *testing.T) {
	tests := []struct {
		x        float64
		expected float64
	}{
		{0.0, 0.0},
		{1.0, 1.0},
		{-1.0, 0.0},
		{10.5, 10.5},
		{-10.5, 0.0},
	}

	for _, tt := range tests {
		result := ReLU(tt.x)
		if math.Abs(result-tt.expected) > 1e-10 {
			t.Errorf("ReLU(%f) = %f; expected %f", tt.x, result, tt.expected)
		}
	}
}

// Test Conversion functions
func TestDegToRad(t *testing.T) {
	tests := []struct {
		degrees  float64
		expected float64
	}{
		{0.0, 0.0},
		{180.0, math.Pi},
		{90.0, math.Pi / 2},
		{360.0, 2 * math.Pi},
		{-180.0, -math.Pi},
	}

	for _, tt := range tests {
		result := DegToRad(tt.degrees)
		if math.Abs(result-tt.expected) > 1e-10 {
			t.Errorf("DegToRad(%f) = %f; expected %f", tt.degrees, result, tt.expected)
		}
	}
}

func TestRadToDeg(t *testing.T) {
	tests := []struct {
		radians  float64
		expected float64
	}{
		{0.0, 0.0},
		{math.Pi, 180.0},
		{math.Pi / 2, 90.0},
		{2 * math.Pi, 360.0},
		{-math.Pi, -180.0},
	}

	for _, tt := range tests {
		result := RadToDeg(tt.radians)
		if math.Abs(result-tt.expected) > 1e-10 {
			t.Errorf("RadToDeg(%f) = %f; expected %f", tt.radians, result, tt.expected)
		}
	}
}

// Test HaversineDistance
func TestHaversineDistance(t *testing.T) {
	tests := []struct {
		lat1, lon1, lat2, lon2 float64
		expected               float64
	}{
		{0.0, 0.0, 0.0, 0.0, 0.0},
		{51.5074, -0.1278, 40.7128, -74.0060, 5570.0}, // London to NYC, approximate
		{90.0, 0.0, -90.0, 0.0, 20015.0},              // North Pole to South Pole
	}

	for _, tt := range tests {
		result := HaversineDistance(tt.lat1, tt.lon1, tt.lat2, tt.lon2)
		// Allow 1% error for Earth curvature calculations
		if math.Abs(result-tt.expected)/tt.expected > 0.01 && tt.expected > 0 {
			t.Errorf("HaversineDistance(%f, %f, %f, %f) = %f; expected ~%f",
				tt.lat1, tt.lon1, tt.lat2, tt.lon2, result, tt.expected)
		}
	}
}

// Test Combinatorics
func TestCombinations(t *testing.T) {
	tests := []struct {
		n, k     int
		expected int
	}{
		{5, 2, 10},
		{10, 3, 120},
		{0, 0, 1},
		{5, 5, 1},
		{5, 0, 1},
		{5, 6, 0},  // k > n
		{5, -1, 0}, // k < 0
	}

	for _, tt := range tests {
		result := Combinations(tt.n, tt.k)
		if result != tt.expected {
			t.Errorf("Combinations(%d, %d) = %d; expected %d", tt.n, tt.k, result, tt.expected)
		}
	}
}

func TestPermutations(t *testing.T) {
	tests := []struct {
		n, k     int
		expected int
	}{
		{5, 2, 20},
		{10, 3, 720},
		{0, 0, 1},
		{5, 5, 120},
		{5, 0, 1},
		{5, 6, 0},  // k > n
		{5, -1, 0}, // k < 0
	}

	for _, tt := range tests {
		result := Permutations(tt.n, tt.k)
		if result != tt.expected {
			t.Errorf("Permutations(%d, %d) = %d; expected %d", tt.n, tt.k, result, tt.expected)
		}
	}
}

// Test Number to words
func TestToWords(t *testing.T) {
	tests := []struct {
		n        int
		expected string
	}{
		{0, "zero"},
		{1, "one"},
		{10, "ten"},
		{25, "twenty-five"},
		{100, "one hundred"},
		{123, "one hundred and twenty-three"},
		{1000, "one thousand"},
		{1234, "one thousand two hundred and thirty-four"},
		{-123, "minus one hundred and twenty-three"},
		{1000000, "one million"},
	}

	for _, tt := range tests {
		result := ToWords(tt.n)
		if result != tt.expected {
			t.Errorf("ToWords(%d) = %s; expected %s", tt.n, result, tt.expected)
		}
	}
}

// Test Roman numeral conversions
func TestRomanToArabic(t *testing.T) {
	tests := []struct {
		roman       string
		expected    int
		expectError bool
	}{
		{"I", 1, false},
		{"IV", 4, false},
		{"V", 5, false},
		{"IX", 9, false},
		{"X", 10, false},
		{"XL", 40, false},
		{"L", 50, false},
		{"XC", 90, false},
		{"C", 100, false},
		{"CD", 400, false},
		{"D", 500, false},
		{"CM", 900, false},
		{"M", 1000, false},
		{"MMMCMXCIX", 3999, false},
		{"i", 1, false}, // lowercase
		{"V", 5, false},
		{"", 0, true},      // empty
		{"ABC", 0, true},   // invalid
		{"IIII", 4, false}, // non-standard but valid
	}

	for _, tt := range tests {
		result, err := RomanToArabic(tt.roman)
		if tt.expectError && err == nil {
			t.Errorf("RomanToArabic(%s) expected error but got none", tt.roman)
		}
		if !tt.expectError && err != nil {
			t.Errorf("RomanToArabic(%s) unexpected error: %v", tt.roman, err)
		}
		if !tt.expectError && result != tt.expected {
			t.Errorf("RomanToArabic(%s) = %d; expected %d", tt.roman, result, tt.expected)
		}
	}
}

func TestArabicToRoman(t *testing.T) {
	tests := []struct {
		n           int
		expected    string
		expectError bool
	}{
		{1, "I", false},
		{4, "IV", false},
		{5, "V", false},
		{9, "IX", false},
		{10, "X", false},
		{40, "XL", false},
		{50, "L", false},
		{90, "XC", false},
		{100, "C", false},
		{400, "CD", false},
		{500, "D", false},
		{900, "CM", false},
		{1000, "M", false},
		{3999, "MMMCMXCIX", false},
		{0, "", true},    // out of range
		{4000, "", true}, // out of range
		{-1, "", true},   // out of range
	}

	for _, tt := range tests {
		result, err := ArabicToRoman(tt.n)
		if tt.expectError && err == nil {
			t.Errorf("ArabicToRoman(%d) expected error but got none", tt.n)
		}
		if !tt.expectError && err != nil {
			t.Errorf("ArabicToRoman(%d) unexpected error: %v", tt.n, err)
		}
		if !tt.expectError && result != tt.expected {
			t.Errorf("ArabicToRoman(%d) = %s; expected %s", tt.n, result, tt.expected)
		}
	}
}

// Helper function for sorting integers
func sortInts(nums []int) {
	sort.Slice(nums, func(i, j int) bool {
		return nums[i] < nums[j]
	})
}

// Benchmark tests
func BenchmarkFibonacci(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Fibonacci(20)
	}
}

func BenchmarkIsPrime(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsPrime(7919) // 7919 is a prime
	}
}

func BenchmarkStandardDeviation(b *testing.B) {
	numbers := GenerateRandomSlice(1000, 1, 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		StandardDeviation(numbers)
	}
}

// Example tests (testable examples)
func ExampleMin() {
	result := Min(5, 10)
	println(result) // Output: 5
}

func ExampleMax() {
	result := Max(5, 10)
	println(result) // Output: 10
}

func ExampleIsPrime() {
	result := IsPrime(17)
	println(result) // Output: true
}
