package main

import (
	"fmt"
	"strings"
)

// Hello prints a greeting message from the sego package
func Hello() {
	fmt.Println("Hello from the Sego package!")
}

// Add takes two integers and returns their sum
func Add(a, b int) int {
	return a + b
}

// Multiply takes two integers and returns their product
func Multiply(a, b int) int {
	return a * b
}

// ToUpper converts a string to uppercase
func ToUpper(s string) string {
	return strings.ToUpper(s)
}

// ReverseString reverses the given string
func ReverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Info prints a custom message with your name
func Info(name string) {
	fmt.Printf("Hello %s! Welcome to Sego package.\n", name)
}
