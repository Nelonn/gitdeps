package main

import (
	"testing"
)

// Test for StrArrContains
func TestStrArrContains(t *testing.T) {
	tests := []struct {
		arr      []string
		str      string
		expected bool
	}{
		{[]string{"a", "b", "c"}, "b", true},
		{[]string{"a", "b", "c"}, "d", false},
		{[]string{}, "a", false},
	}

	for _, test := range tests {
		result := StrArrContains(test.arr, test.str)
		if result != test.expected {
			t.Errorf("StrArrContains(%v, %v) = %v; want %v", test.arr, test.str, result, test.expected)
		}
	}
}

// Test for StrArrMoreThanOneNotEmpty
func TestStrArrMoreThanOneNotEmpty(t *testing.T) {
	tests := []struct {
		arr      []string
		expected bool
	}{
		{[]string{"a", "b", ""}, true},
		{[]string{"a", "", ""}, false},
		{[]string{"", "", ""}, false},
		{[]string{"a", "b", "c"}, true},
	}

	for _, test := range tests {
		result := StrArrMoreThanOneNotEmpty(test.arr)
		if result != test.expected {
			t.Errorf("StrArrMoreThanOneNotEmpty(%v) = %v; want %v", test.arr, result, test.expected)
		}
	}
}

// Test for CheckStrDuplicates
func TestCheckStrDuplicates(t *testing.T) {
	tests := []struct {
		arr      []string
		expected string
	}{
		{[]string{"a", "b", "c"}, ""},
		{[]string{"a", "b", "a"}, "a"},
		{[]string{"a", "b", "b"}, "b"},
		{[]string{"a", "a", "a"}, "a"},
	}

	for _, test := range tests {
		result := CheckStrDuplicates(test.arr)
		if result != test.expected {
			t.Errorf("CheckStrDuplicates(%v) = %v; want %v", test.arr, result, test.expected)
		}
	}
}
