package words

import (
	"strings"
	"unicode"

	"github.com/kljensen/snowball"
)

func Normalize(in string) []string {
	result := make([]string, 0)
	words := strings.FieldsFunc(in, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	seen := make(map[string]bool)
	for _, word := range words {
		word = strings.ToLower(word)
		stemmed, err := snowball.Stem(word, "english", true)
		if err != nil {
			stemmed = word
		}
		if _, ok := stopWords[word]; ok {
			continue
		}

		if !seen[stemmed] && stemmed != "" {
			seen[stemmed] = true
			result = append(result, stemmed)
		}
	}
	return result
}

var stopWords = map[string]bool{
	"a": true, "an": true, "the": true,
	"i": true, "he": true, "she": true, "it": true,
	"that": true,
	"and":  true,
	"you":  true,
	"your": true,
	"or":   true,
	"me":   true,
	"them": true,
	"who":  true,
	"will": true,
}
