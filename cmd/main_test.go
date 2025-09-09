package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcat(t *testing.T) {
	// Test case 1: Normal strings
	result := Concat("Hello", "World")
	assert.Equal(t, "Hello World", result, "Concat should join strings with a space")

	// Test case 2: Empty strings
	result = Concat("", "")
	assert.Equal(t, " ", result, "Concat should return a space for empty strings")
}
