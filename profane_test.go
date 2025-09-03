package main

import "testing"

func TestProfaneCheck(t *testing.T) {
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}
	cases := []struct {
		input    string
		expected string
	}{
		{
			input:    "This is a kerfuffle opinion I need to share with the world",
			expected: "This is a **** opinion I need to share with the world",
		},
		{
			input:    "I had something interesting for breakfast",
			expected: "I had something interesting for breakfast",
		},
	}

	for _, c := range cases {
		actual := profaneCheck(c.input, profaneWords)
		if c.expected != actual {
			t.Errorf("Expected: %q, got: %q", c.expected, actual)
		}
	}
}
