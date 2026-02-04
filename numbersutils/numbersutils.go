package numbersutils

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// Min returns the minimum of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MinInt64 returns the minimum of two int64
func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// MinFloat64 returns the minimum of two float64
func MinFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of two integers
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// MaxInt64 returns the maximum of two int64
func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// MaxFloat64 returns the maximum of two float64
func MaxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// Abs returns the absolute value of an integer
func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// AbsInt64 returns the absolute value of an int64
func AbsInt64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

// AbsFloat64 returns the absolute value of a float64
func AbsFloat64(x float64) float64 {
	return math.Abs(x)
}

// Sum returns the sum of a slice of integers
func Sum(numbers []int) int {
	total := 0
	for _, num := range numbers {
		total += num
	}
	return total
}

// SumInt64 returns the sum of a slice of int64
func SumInt64(numbers []int64) int64 {
	var total int64 = 0
	for _, num := range numbers {
		total += num
	}
	return total
}

// SumFloat64 returns the sum of a slice of float64
func SumFloat64(numbers []float64) float64 {
	var total float64 = 0
	for _, num := range numbers {
		total += num
	}
	return total
}

// Average returns the average of a slice of integers as float64
func Average(numbers []int) float64 {
	if len(numbers) == 0 {
		return 0
	}
	return float64(Sum(numbers)) / float64(len(numbers))
}

// AverageInt64 returns the average of a slice of int64 as float64
func AverageInt64(numbers []int64) float64 {
	if len(numbers) == 0 {
		return 0
	}
	return float64(SumInt64(numbers)) / float64(len(numbers))
}

// AverageFloat64 returns the average of a slice of float64
func AverageFloat64(numbers []float64) float64 {
	if len(numbers) == 0 {
		return 0
	}
	return SumFloat64(numbers) / float64(len(numbers))
}

// Median returns the median of a slice of integers
func Median(numbers []int) float64 {
	if len(numbers) == 0 {
		return 0
	}

	// Create a copy to avoid modifying original
	sorted := make([]int, len(numbers))
	copy(sorted, numbers)
	sort.Ints(sorted)

	n := len(sorted)
	if n%2 == 1 {
		// Odd number of elements
		return float64(sorted[n/2])
	}

	// Even number of elements
	return float64(sorted[n/2-1]+sorted[n/2]) / 2.0
}

// MedianFloat64 returns the median of a slice of float64
func MedianFloat64(numbers []float64) float64 {
	if len(numbers) == 0 {
		return 0
	}

	// Create a copy to avoid modifying original
	sorted := make([]float64, len(numbers))
	copy(sorted, numbers)
	sort.Float64s(sorted)

	n := len(sorted)
	if n%2 == 1 {
		// Odd number of elements
		return sorted[n/2]
	}

	// Even number of elements
	return (sorted[n/2-1] + sorted[n/2]) / 2.0
}

// Mode returns the mode(s) of a slice of integers
func Mode(numbers []int) []int {
	if len(numbers) == 0 {
		return []int{}
	}

	frequency := make(map[int]int)
	for _, num := range numbers {
		frequency[num]++
	}

	maxFreq := 0
	for _, freq := range frequency {
		if freq > maxFreq {
			maxFreq = freq
		}
	}

	var modes []int
	for num, freq := range frequency {
		if freq == maxFreq {
			modes = append(modes, num)
		}
	}

	sort.Ints(modes)
	return modes
}

// StandardDeviation returns the standard deviation of a slice of integers
func StandardDeviation(numbers []int) float64 {
	if len(numbers) < 2 {
		return 0
	}

	mean := Average(numbers)
	var sumSquares float64

	for _, num := range numbers {
		diff := float64(num) - mean
		sumSquares += diff * diff
	}

	variance := sumSquares / float64(len(numbers)-1)
	return math.Sqrt(variance)
}

// StandardDeviationFloat64 returns the standard deviation of a slice of float64
func StandardDeviationFloat64(numbers []float64) float64 {
	if len(numbers) < 2 {
		return 0
	}

	mean := AverageFloat64(numbers)
	var sumSquares float64

	for _, num := range numbers {
		diff := num - mean
		sumSquares += diff * diff
	}

	variance := sumSquares / float64(len(numbers)-1)
	return math.Sqrt(variance)
}

// IsEven checks if a number is even
func IsEven(n int) bool {
	return n%2 == 0
}

// IsOdd checks if a number is odd
func IsOdd(n int) bool {
	return n%2 != 0
}

// IsPrime checks if a number is prime
func IsPrime(n int) bool {
	if n <= 1 {
		return false
	}
	if n == 2 {
		return true
	}
	if n%2 == 0 {
		return false
	}

	sqrt := int(math.Sqrt(float64(n)))
	for i := 3; i <= sqrt; i += 2 {
		if n%i == 0 {
			return false
		}
	}
	return true
}

// IsPrimeBig checks if a big integer is prime (for large numbers)
func IsPrimeBig(n *big.Int) bool {
	return n.ProbablyPrime(20) // 20 iterations for high confidence
}

// Factorial calculates the factorial of a number
func Factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n * Factorial(n-1)
}

// FactorialBig calculates factorial for large numbers
func FactorialBig(n int64) *big.Int {
	if n <= 1 {
		return big.NewInt(1)
	}

	result := big.NewInt(1)
	for i := int64(2); i <= n; i++ {
		result.Mul(result, big.NewInt(i))
	}
	return result
}

// Fibonacci returns the nth Fibonacci number
func Fibonacci(n int) int {
	if n <= 1 {
		return n
	}

	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

// FibonacciSequence returns first n Fibonacci numbers
func FibonacciSequence(n int) []int {
	if n <= 0 {
		return []int{}
	}

	sequence := make([]int, n)
	if n >= 1 {
		sequence[0] = 0
	}
	if n >= 2 {
		sequence[1] = 1
	}

	for i := 2; i < n; i++ {
		sequence[i] = sequence[i-1] + sequence[i-2]
	}

	return sequence
}

// GenerateRandomInt generates a random integer within a range [min, max]
func GenerateRandomInt(min, max int) int {
	if min > max {
		min, max = max, min
	}

	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min+1) + min
}

// GenerateRandomInt64 generates a random int64 within a range [min, max]
func GenerateRandomInt64(min, max int64) int64 {
	if min > max {
		min, max = max, min
	}

	rand.Seed(time.Now().UnixNano())
	return min + rand.Int63n(max-min+1)
}

// GenerateRandomFloat64 generates a random float64 within a range [min, max]
func GenerateRandomFloat64(min, max float64) float64 {
	if min > max {
		min, max = max, min
	}

	rand.Seed(time.Now().UnixNano())
	return min + rand.Float64()*(max-min)
}

// GenerateRandomSlice generates a slice of random integers
func GenerateRandomSlice(size, min, max int) []int {
	if size <= 0 {
		return []int{}
	}

	result := make([]int, size)
	for i := 0; i < size; i++ {
		result[i] = GenerateRandomInt(min, max)
	}
	return result
}

// Round rounds a float64 to specified decimal places
func Round(value float64, decimals int) float64 {
	if decimals < 0 {
		decimals = 0
	}

	multiplier := math.Pow(10, float64(decimals))
	return math.Round(value*multiplier) / multiplier
}

// RoundUp rounds a float64 up to specified decimal places
func RoundUp(value float64, decimals int) float64 {
	if decimals < 0 {
		decimals = 0
	}

	multiplier := math.Pow(10, float64(decimals))
	return math.Ceil(value*multiplier) / multiplier
}

// RoundDown rounds a float64 down to specified decimal places
func RoundDown(value float64, decimals int) float64 {
	if decimals < 0 {
		decimals = 0
	}

	multiplier := math.Pow(10, float64(decimals))
	return math.Floor(value*multiplier) / multiplier
}

// FormatCurrency formats a number as currency
func FormatCurrency(amount float64, currencySymbol string, decimals int) string {
	rounded := Round(amount, decimals)
	formatted := fmt.Sprintf("%s%.*f", currencySymbol, decimals, rounded)

	// Add thousand separators
	parts := strings.Split(formatted, ".")
	integerPart := parts[0]

	// Handle negative numbers
	negative := false
	if strings.HasPrefix(integerPart, "-") {
		negative = true
		integerPart = integerPart[1:]
	}
	if currencySymbol != "" && strings.HasPrefix(integerPart, currencySymbol) {
		integerPart = strings.TrimPrefix(integerPart, currencySymbol)
	}

	// Add commas
	var result strings.Builder
	count := 0
	for i := len(integerPart) - 1; i >= 0; i-- {
		if count == 3 {
			result.WriteString(",")
			count = 0
		}
		result.WriteByte(integerPart[i])
		count++
	}

	// Reverse and add back sign and symbol
	integerWithCommas := reverseString(result.String())

	if negative {
		integerWithCommas = "-" + integerWithCommas
	}

	if currencySymbol != "" {
		integerWithCommas = currencySymbol + integerWithCommas
	}

	// Add decimal part if needed
	if len(parts) > 1 {
		return integerWithCommas + "." + parts[1]
	}

	return integerWithCommas
}

// FormatNumber formats a number with thousand separators
func FormatNumber(number float64, decimals int) string {
	return FormatCurrency(number, "", decimals)
}

// ParseFloatSafe safely parses a string to float64
func ParseFloatSafe(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty string")
	}
	return strconv.ParseFloat(s, 64)
}

// ParseIntSafe safely parses a string to int
func ParseIntSafe(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty string")
	}
	return strconv.Atoi(s)
}

// ParseInt64Safe safely parses a string to int64
func ParseInt64Safe(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty string")
	}
	return strconv.ParseInt(s, 10, 64)
}

// ParseBoolSafe safely parses a string to bool
func ParseBoolSafe(s string) (bool, error) {
	s = strings.TrimSpace(strings.ToLower(s))

	switch s {
	case "true", "t", "yes", "y", "1", "on":
		return true, nil
	case "false", "f", "no", "n", "0", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean string: %s", s)
	}
}

// Clamp clamps a value between min and max
func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// ClampFloat64 clamps a float64 value between min and max
func ClampFloat64(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// IsBetween checks if a number is between two values (inclusive)
func IsBetween(value, min, max int) bool {
	return value >= min && value <= max
}

// IsBetweenFloat64 checks if a float64 is between two values (inclusive)
func IsBetweenFloat64(value, min, max float64) bool {
	return value >= min && value <= max
}

// GCD calculates the Greatest Common Divisor of two numbers (Euclidean algorithm)
func GCD(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return Abs(a)
}

// GCDInt64 calculates the Greatest Common Divisor of two int64 numbers
func GCDInt64(a, b int64) int64 {
	for b != 0 {
		a, b = b, a%b
	}
	return AbsInt64(a)
}

// LCM calculates the Least Common Multiple of two numbers
func LCM(a, b int) int {
	if a == 0 || b == 0 {
		return 0
	}
	return Abs(a*b) / GCD(a, b)
}

// LCMInt64 calculates the Least Common Multiple of two int64 numbers
func LCMInt64(a, b int64) int64 {
	if a == 0 || b == 0 {
		return 0
	}
	return AbsInt64(a*b) / GCDInt64(a, b)
}

// PercentChange calculates the percentage change between two values
func PercentChange(oldValue, newValue float64) float64 {
	if oldValue == 0 {
		if newValue == 0 {
			return 0
		}
		return 100 // From 0 to non-zero is 100% increase
	}
	return ((newValue - oldValue) / oldValue) * 100
}

// PercentOf calculates what percent a is of b
func PercentOf(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return (a / b) * 100
}

// PercentValue calculates the value of a percentage
func PercentValue(percent, total float64) float64 {
	return (percent / 100) * total
}

// IsPerfectSquare checks if a number is a perfect square
func IsPerfectSquare(n int) bool {
	if n < 0 {
		return false
	}
	sqrt := int(math.Sqrt(float64(n)))
	return sqrt*sqrt == n
}

// IsPowerOfTwo checks if a number is a power of two
func IsPowerOfTwo(n int) bool {
	if n <= 0 {
		return false
	}
	return (n & (n - 1)) == 0
}

// IsArmstrong checks if a number is an Armstrong number (narcissistic number)
func IsArmstrong(num int) bool {
	if num < 0 {
		return false
	}

	original := num
	sum := 0
	digits := 0

	// Count digits
	temp := num
	for temp > 0 {
		temp /= 10
		digits++
	}

	// Calculate sum of digits^digitCount
	temp = num
	for temp > 0 {
		digit := temp % 10
		sum += int(math.Pow(float64(digit), float64(digits)))
		temp /= 10
	}

	return sum == original
}

// IsPalindromeNumber checks if a number is a palindrome
func IsPalindromeNumber(num int) bool {
	if num < 0 {
		return false
	}

	original := num
	reversed := 0

	for num > 0 {
		digit := num % 10
		reversed = reversed*10 + digit
		num /= 10
	}

	return original == reversed
}

// Digits returns the digits of a number as a slice
func Digits(num int) []int {
	if num == 0 {
		return []int{0}
	}

	if num < 0 {
		num = -num
	}

	var digits []int
	for num > 0 {
		digits = append([]int{num % 10}, digits...)
		num /= 10
	}

	return digits
}

// DigitSum returns the sum of digits of a number
func DigitSum(num int) int {
	sum := 0
	for num != 0 {
		sum += num % 10
		num /= 10
	}
	return sum
}

// ReverseNumber reverses the digits of a number
func ReverseNumber(num int) int {
	reversed := 0
	for num != 0 {
		reversed = reversed*10 + num%10
		num /= 10
	}
	return reversed
}

// BaseConvert converts a number from one base to another (2-36)
func BaseConvert(num string, fromBase, toBase int) (string, error) {
	if fromBase < 2 || fromBase > 36 || toBase < 2 || toBase > 36 {
		return "", errors.New("base must be between 2 and 36")
	}

	// Parse the number from the source base
	n := new(big.Int)
	_, success := n.SetString(num, fromBase)
	if !success {
		return "", errors.New("invalid number for the given base")
	}

	// Convert to the target base
	return n.Text(toBase), nil
}

// ToBinary converts a decimal number to binary string
func ToBinary(n int) string {
	return strconv.FormatInt(int64(n), 2)
}

// ToHex converts a decimal number to hexadecimal string
func ToHex(n int) string {
	return strconv.FormatInt(int64(n), 16)
}

// ToOctal converts a decimal number to octal string
func ToOctal(n int) string {
	return strconv.FormatInt(int64(n), 8)
}

// FromBinary converts a binary string to decimal
func FromBinary(bin string) (int, error) {
	n, err := strconv.ParseInt(bin, 2, 64)
	return int(n), err
}

// FromHex converts a hexadecimal string to decimal
func FromHex(hex string) (int, error) {
	n, err := strconv.ParseInt(hex, 16, 64)
	return int(n), err
}

// FromOctal converts an octal string to decimal
func FromOctal(oct string) (int, error) {
	n, err := strconv.ParseInt(oct, 8, 64)
	return int(n), err
}

// IsValidCreditCard checks if a string is a valid credit card number using Luhn algorithm
func IsValidCreditCard(number string) bool {
	// Remove all non-digit characters
	re := regexp.MustCompile(`\D`)
	clean := re.ReplaceAllString(number, "")

	if len(clean) < 13 || len(clean) > 19 {
		return false
	}

	// Luhn algorithm
	sum := 0
	double := false

	for i := len(clean) - 1; i >= 0; i-- {
		digit := int(clean[i] - '0')

		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		double = !double
	}

	return sum%10 == 0
}

// IsValidISBN checks if a string is a valid ISBN-10 or ISBN-13
func IsValidISBN(isbn string) bool {
	// Remove hyphens and spaces
	re := regexp.MustCompile(`[-\s]`)
	clean := re.ReplaceAllString(isbn, "")

	if len(clean) == 10 {
		return isValidISBN10(clean)
	} else if len(clean) == 13 {
		return isValidISBN13(clean)
	}

	return false
}

func isValidISBN10(isbn string) bool {
	if len(isbn) != 10 {
		return false
	}

	sum := 0
	for i := 0; i < 9; i++ {
		digit := int(isbn[i] - '0')
		if digit < 0 || digit > 9 {
			return false
		}
		sum += digit * (10 - i)
	}

	lastChar := isbn[9]
	if lastChar == 'X' || lastChar == 'x' {
		sum += 10
	} else {
		digit := int(lastChar - '0')
		if digit < 0 || digit > 9 {
			return false
		}
		sum += digit
	}

	return sum%11 == 0
}

func isValidISBN13(isbn string) bool {
	if len(isbn) != 13 {
		return false
	}

	sum := 0
	for i := 0; i < 13; i++ {
		digit := int(isbn[i] - '0')
		if digit < 0 || digit > 9 {
			return false
		}

		if i%2 == 0 {
			sum += digit
		} else {
			sum += digit * 3
		}
	}

	return sum%10 == 0
}

// IsValidPhoneNumber checks if a string is a valid phone number (basic validation)
func IsValidPhoneNumber(phone string) bool {
	// Remove all non-digit characters
	re := regexp.MustCompile(`\D`)
	clean := re.ReplaceAllString(phone, "")

	// Basic validation: between 7 and 15 digits
	return len(clean) >= 7 && len(clean) <= 15
}

// IsValidPostalCode checks if a string is a valid postal code (US format)
func IsValidPostalCode(code string) bool {
	// US ZIP code format: 5 digits or 5+4 digits
	re := regexp.MustCompile(`^\d{5}(-\d{4})?$`)
	return re.MatchString(code)
}

// IsValidSSN checks if a string is a valid US Social Security Number
func IsValidSSN(ssn string) bool {
	// Format: XXX-XX-XXXX
	re := regexp.MustCompile(`^\d{3}-\d{2}-\d{4}$`)
	if !re.MatchString(ssn) {
		return false
	}

	// Remove hyphens
	clean := strings.ReplaceAll(ssn, "-", "")

	// Check for invalid numbers
	if clean == "000000000" ||
		clean == "123456789" ||
		clean[0:3] == "000" ||
		clean[3:5] == "00" ||
		clean[5:9] == "0000" {
		return false
	}

	return true
}

// RandomChoice randomly selects n elements from a slice
func RandomChoice(numbers []int, n int) []int {
	if n <= 0 || len(numbers) == 0 {
		return []int{}
	}

	if n >= len(numbers) {
		// Return shuffled copy of all elements
		shuffled := make([]int, len(numbers))
		copy(shuffled, numbers)
		rand.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})
		return shuffled
	}

	rand.Seed(time.Now().UnixNano())
	selected := make([]int, n)
	indices := rand.Perm(len(numbers))

	for i := 0; i < n; i++ {
		selected[i] = numbers[indices[i]]
	}

	return selected
}

// Shuffle randomly shuffles a slice of integers
func Shuffle(numbers []int) []int {
	if len(numbers) == 0 {
		return []int{}
	}

	shuffled := make([]int, len(numbers))
	copy(shuffled, numbers)

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled
}

// Unique returns unique integers from a slice
func Unique(numbers []int) []int {
	if len(numbers) == 0 {
		return []int{}
	}

	seen := make(map[int]bool)
	var result []int

	for _, num := range numbers {
		if !seen[num] {
			seen[num] = true
			result = append(result, num)
		}
	}

	return result
}

// Intersection returns common elements between two slices
func Intersection(a, b []int) []int {
	setA := make(map[int]bool)
	for _, num := range a {
		setA[num] = true
	}

	var result []int
	for _, num := range b {
		if setA[num] {
			result = append(result, num)
			setA[num] = false // Avoid duplicates
		}
	}

	return result
}

// Union returns all unique elements from both slices
func Union(a, b []int) []int {
	set := make(map[int]bool)

	for _, num := range a {
		set[num] = true
	}

	for _, num := range b {
		set[num] = true
	}

	result := make([]int, 0, len(set))
	for num := range set {
		result = append(result, num)
	}

	return result
}

// Difference returns elements in a that are not in b
func Difference(a, b []int) []int {
	setB := make(map[int]bool)
	for _, num := range b {
		setB[num] = true
	}

	var result []int
	for _, num := range a {
		if !setB[num] {
			result = append(result, num)
		}
	}

	return result
}

// Range generates a sequence of numbers from start to end (exclusive)
func Range(start, end, step int) []int {
	if step == 0 {
		return []int{}
	}

	if (step > 0 && start >= end) || (step < 0 && start <= end) {
		return []int{}
	}

	var result []int
	for i := start; (step > 0 && i < end) || (step < 0 && i > end); i += step {
		result = append(result, i)
	}

	return result
}

// RangeInclusive generates a sequence of numbers from start to end (inclusive)
func RangeInclusive(start, end, step int) []int {
	if step == 0 {
		return []int{}
	}

	if (step > 0 && start > end) || (step < 0 && start < end) {
		return []int{}
	}

	var result []int
	for i := start; (step > 0 && i <= end) || (step < 0 && i >= end); i += step {
		result = append(result, i)
	}

	return result
}

// Linspace generates n evenly spaced numbers between start and end (inclusive)
func Linspace(start, end float64, n int) []float64 {
	if n <= 0 {
		return []float64{}
	}
	if n == 1 {
		return []float64{start}
	}

	result := make([]float64, n)
	step := (end - start) / float64(n-1)

	for i := 0; i < n; i++ {
		result[i] = start + float64(i)*step
	}

	return result
}

// Arange generates numbers from start to end with step (like numpy.arange)
func Arange(start, end, step float64) []float64 {
	if step == 0 {
		return []float64{}
	}

	if (step > 0 && start >= end) || (step < 0 && start <= end) {
		return []float64{}
	}

	var result []float64
	for i := start; (step > 0 && i < end) || (step < 0 && i > end); i += step {
		result = append(result, i)
	}

	return result
}

// Normalize scales numbers to range [0, 1]
func Normalize(numbers []float64) []float64 {
	if len(numbers) == 0 {
		return []float64{}
	}

	// Find min and max
	min := numbers[0]
	max := numbers[0]

	for _, num := range numbers {
		if num < min {
			min = num
		}
		if num > max {
			max = num
		}
	}

	// All numbers are the same
	if max == min {
		return make([]float64, len(numbers))
	}

	// Normalize
	result := make([]float64, len(numbers))
	for i, num := range numbers {
		result[i] = (num - min) / (max - min)
	}

	return result
}

// Standardize scales numbers to have mean 0 and standard deviation 1
func Standardize(numbers []float64) []float64 {
	if len(numbers) < 2 {
		return make([]float64, len(numbers))
	}

	mean := AverageFloat64(numbers)
	std := StandardDeviationFloat64(numbers)

	if std == 0 {
		return make([]float64, len(numbers))
	}

	result := make([]float64, len(numbers))
	for i, num := range numbers {
		result[i] = (num - mean) / std
	}

	return result
}

// Quantile returns the value at a given quantile (0 to 1)
func Quantile(numbers []float64, q float64) float64 {
	if len(numbers) == 0 || q < 0 || q > 1 {
		return 0
	}

	sorted := make([]float64, len(numbers))
	copy(sorted, numbers)
	sort.Float64s(sorted)

	index := q * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sorted[lower]
	}

	// Linear interpolation
	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// Quartiles returns the three quartiles (Q1, Q2, Q3)
func Quartiles(numbers []float64) (float64, float64, float64) {
	if len(numbers) == 0 {
		return 0, 0, 0
	}

	return Quantile(numbers, 0.25), Quantile(numbers, 0.5), Quantile(numbers, 0.75)
}

// IQR returns the interquartile range (Q3 - Q1)
func IQR(numbers []float64) float64 {
	if len(numbers) == 0 {
		return 0
	}

	q1, _, q3 := Quartiles(numbers)
	return q3 - q1
}

// IsOutlier checks if a value is an outlier using IQR method
func IsOutlier(value float64, numbers []float64) bool {
	if len(numbers) < 4 {
		return false
	}

	q1, _, q3 := Quartiles(numbers)
	iqr := q3 - q1
	lowerBound := q1 - 1.5*iqr
	upperBound := q3 + 1.5*iqr

	return value < lowerBound || value > upperBound
}

// MovingAverage calculates moving average of a slice
func MovingAverage(numbers []float64, windowSize int) []float64 {
	if len(numbers) == 0 || windowSize <= 0 || windowSize > len(numbers) {
		return []float64{}
	}

	result := make([]float64, len(numbers)-windowSize+1)

	for i := 0; i <= len(numbers)-windowSize; i++ {
		sum := 0.0
		for j := 0; j < windowSize; j++ {
			sum += numbers[i+j]
		}
		result[i] = sum / float64(windowSize)
	}

	return result
}

// CumulativeSum calculates cumulative sum of a slice
func CumulativeSum(numbers []float64) []float64 {
	if len(numbers) == 0 {
		return []float64{}
	}

	result := make([]float64, len(numbers))
	result[0] = numbers[0]

	for i := 1; i < len(numbers); i++ {
		result[i] = result[i-1] + numbers[i]
	}

	return result
}

// DotProduct calculates dot product of two slices
func DotProduct(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	sum := 0.0
	for i := 0; i < len(a); i++ {
		sum += a[i] * b[i]
	}

	return sum
}

// EuclideanDistance calculates Euclidean distance between two vectors
func EuclideanDistance(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	sum := 0.0
	for i := 0; i < len(a); i++ {
		diff := a[i] - b[i]
		sum += diff * diff
	}

	return math.Sqrt(sum)
}

// ManhattanDistance calculates Manhattan distance between two vectors
func ManhattanDistance(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	sum := 0.0
	for i := 0; i < len(a); i++ {
		sum += math.Abs(a[i] - b[i])
	}

	return sum
}

// CosineSimilarity calculates cosine similarity between two vectors
func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	dot := DotProduct(a, b)

	normA := math.Sqrt(DotProduct(a, a))
	normB := math.Sqrt(DotProduct(b, b))

	if normA == 0 || normB == 0 {
		return 0
	}

	return dot / (normA * normB)
}

// Correlation calculates Pearson correlation coefficient
func Correlation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0
	}

	n := float64(len(x))

	// Calculate means
	meanX := AverageFloat64(x)
	meanY := AverageFloat64(y)

	// Calculate covariance and standard deviations
	var cov, stdX, stdY float64

	for i := 0; i < len(x); i++ {
		dx := x[i] - meanX
		dy := y[i] - meanY

		cov += dx * dy
		stdX += dx * dx
		stdY += dy * dy
	}

	cov /= n - 1
	stdX = math.Sqrt(stdX / (n - 1))
	stdY = math.Sqrt(stdY / (n - 1))

	if stdX == 0 || stdY == 0 {
		return 0
	}

	return cov / (stdX * stdY)
}

// Sigmoid calculates sigmoid function
func Sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

// Softmax calculates softmax function for a slice
func Softmax(numbers []float64) []float64 {
	if len(numbers) == 0 {
		return []float64{}
	}

	// Find max for numerical stability
	max := numbers[0]
	for _, num := range numbers {
		if num > max {
			max = num
		}
	}

	// Calculate exponentials
	expSum := 0.0
	exps := make([]float64, len(numbers))

	for i, num := range numbers {
		exps[i] = math.Exp(num - max)
		expSum += exps[i]
	}

	// Normalize
	result := make([]float64, len(numbers))
	for i, exp := range exps {
		result[i] = exp / expSum
	}

	return result
}

// ReLU calculates ReLU function
func ReLU(x float64) float64 {
	if x > 0 {
		return x
	}
	return 0
}

// LeakyReLU calculates Leaky ReLU function
func LeakyReLU(x, alpha float64) float64 {
	if x > 0 {
		return x
	}
	return alpha * x
}

// Tanh calculates hyperbolic tangent
func Tanh(x float64) float64 {
	return math.Tanh(x)
}

// DegToRad converts degrees to radians
func DegToRad(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}

// RadToDeg converts radians to degrees
func RadToDeg(radians float64) float64 {
	return radians * 180.0 / math.Pi
}

// HaversineDistance calculates distance between two points on Earth using Haversine formula
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert to radians
	lat1Rad := DegToRad(lat1)
	lon1Rad := DegToRad(lon1)
	lat2Rad := DegToRad(lat2)
	lon2Rad := DegToRad(lon2)

	// Haversine formula
	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Earth radius in kilometers
	R := 6371.0

	return R * c
}

// Combinations calculates number of combinations C(n, k)
func Combinations(n, k int) int {
	if k < 0 || k > n {
		return 0
	}

	if k > n/2 {
		k = n - k
	}

	result := 1
	for i := 1; i <= k; i++ {
		result = result * (n - k + i) / i
	}

	return result
}

// Permutations calculates number of permutations P(n, k)
func Permutations(n, k int) int {
	if k < 0 || k > n {
		return 0
	}

	result := 1
	for i := 0; i < k; i++ {
		result *= n - i
	}

	return result
}

// Entropy calculates Shannon entropy of a probability distribution
func Entropy(probabilities []float64) float64 {
	if len(probabilities) == 0 {
		return 0
	}

	entropy := 0.0
	for _, p := range probabilities {
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}

	return entropy
}

// JaccardSimilarity calculates Jaccard similarity between two sets
func JaccardSimilarity(a, b []int) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}

	setA := make(map[int]bool)
	for _, val := range a {
		setA[val] = true
	}

	intersection := 0
	for _, val := range b {
		if setA[val] {
			intersection++
		}
	}

	union := len(setA) + len(b) - intersection

	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}

// ToWords converts a number to words (English)
func ToWords(n int) string {
	if n == 0 {
		return "zero"
	}

	if n < 0 {
		return "minus " + ToWords(-n)
	}

	units := []string{"", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine"}
	teens := []string{"ten", "eleven", "twelve", "thirteen", "fourteen", "fifteen", "sixteen", "seventeen", "eighteen", "nineteen"}
	tens := []string{"", "", "twenty", "thirty", "forty", "fifty", "sixty", "seventy", "eighty", "ninety"}
	thousands := []string{"", "thousand", "million", "billion", "trillion"}

	var words []string

	// Handle groups of three digits
	groupIndex := 0
	for n > 0 {
		group := n % 1000
		if group > 0 {
			groupWords := convertThreeDigitGroup(group, units, teens, tens)
			if groupIndex > 0 {
				groupWords += " " + thousands[groupIndex]
			}
			words = append([]string{groupWords}, words...)
		}
		n /= 1000
		groupIndex++
	}

	return strings.Join(words, " ")
}

func convertThreeDigitGroup(n int, units, teens, tens []string) string {
	var parts []string

	hundreds := n / 100
	if hundreds > 0 {
		parts = append(parts, units[hundreds]+" hundred")
		n %= 100
	}

	if n > 0 {
		if n < 10 {
			parts = append(parts, units[n])
		} else if n < 20 {
			parts = append(parts, teens[n-10])
		} else {
			tensDigit := n / 10
			unitsDigit := n % 10
			if unitsDigit > 0 {
				parts = append(parts, tens[tensDigit]+"-"+units[unitsDigit])
			} else {
				parts = append(parts, tens[tensDigit])
			}
		}
	}

	return strings.Join(parts, " and ")
}

// RomanToArabic converts Roman numeral to Arabic number
func RomanToArabic(roman string) (int, error) {
	values := map[byte]int{
		'I': 1,
		'V': 5,
		'X': 10,
		'L': 50,
		'C': 100,
		'D': 500,
		'M': 1000,
	}

	if len(roman) == 0 {
		return 0, errors.New("empty roman numeral")
	}

	total := 0
	prevValue := 0

	for i := len(roman) - 1; i >= 0; i-- {
		ch := roman[i]
		if !unicode.IsUpper(rune(ch)) {
			ch = byte(unicode.ToUpper(rune(ch)))
		}

		value, ok := values[ch]
		if !ok {
			return 0, fmt.Errorf("invalid roman numeral character: %c", ch)
		}

		if value < prevValue {
			total -= value
		} else {
			total += value
		}

		prevValue = value
	}

	return total, nil
}

// ArabicToRoman converts Arabic number to Roman numeral
func ArabicToRoman(n int) (string, error) {
	if n <= 0 || n >= 4000 {
		return "", errors.New("number must be between 1 and 3999")
	}

	values := []struct {
		value  int
		symbol string
	}{
		{1000, "M"},
		{900, "CM"},
		{500, "D"},
		{400, "CD"},
		{100, "C"},
		{90, "XC"},
		{50, "L"},
		{40, "XL"},
		{10, "X"},
		{9, "IX"},
		{5, "V"},
		{4, "IV"},
		{1, "I"},
	}

	var result strings.Builder
	remaining := n

	for _, v := range values {
		for remaining >= v.value {
			result.WriteString(v.symbol)
			remaining -= v.value
		}
	}

	return result.String(), nil
}

// Helper function used by FormatCurrency
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
