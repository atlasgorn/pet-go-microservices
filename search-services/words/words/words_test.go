package words

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "single word",
			input:    "simple",
			expected: []string{"simpl"},
		},
		{
			name:     "multiple words with stemming",
			input:    "I follow followers",
			expected: []string{"follow"},
		},
		{
			name:     "punctuation removal",
			input:    "I shouted: 'give me your car!!!",
			expected: []string{"shout", "give", "car"},
		},
		{
			name:     "stop words only",
			input:    "I and you or me or them, who will?",
			expected: []string{},
		},
		{
			name:     "mixed alphanumeric",
			input:    "Moscow!123'check-it'or   123, man,that,difficult:heck",
			expected: []string{"moscow", "check", "123", "man", "difficult", "heck"},
		},
		{
			name:     "duplicate words deduplication",
			input:    "test test TEST testing",
			expected: []string{"test"},
		},
		{
			name:     "numbers preserved",
			input:    "version 2.0 release 2024",
			expected: []string{"version", "2", "0", "releas", "2024"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Normalize(tt.input)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestNormalizeCaseHandling(t *testing.T) {
	result := Normalize("HELLO World HeLLo")
	assert.ElementsMatch(t, []string{"hello", "world"}, result)
}

func TestNormalizeStopWords(t *testing.T) {
	stopPhrase := "the a an and or but in on at to for of with by from"
	result := Normalize(stopPhrase)
	assert.Empty(t, result)
}
