package main

import (
	"reflect"
	"testing"
)

func TestCleanInput(t *testing.T) {
	type test struct {
		input    string
		expected []string
	}

	tests := []test{
		{input: "  hello  world  ", expected: []string{"hello", "world"}},
		{input: "Does, this work?", expected: []string{"does,", "this", "work?"}},
		{input: "does,\nthis\twork?", expected: []string{"does,", "this", "work?"}},
	}

	for _, tc := range tests {
		actual := cleanInput(tc.input)
		if !reflect.DeepEqual(actual, tc.expected) {
			t.Fatalf("Expected: %v, got: %v", tc.expected, actual)
		}
	}
}
