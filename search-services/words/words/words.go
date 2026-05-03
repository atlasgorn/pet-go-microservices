package words

import (
	"strings"
	"unicode"

	"github.com/kljensen/snowball"
	"github.com/kljensen/snowball/english"
)

func Normalize(in string) []string {
	var result []string
	words := strings.FieldsFunc(in, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	seen := make(map[string]bool)
	for _, word := range words {
		word = strings.ToLower(word)
		if english.IsStopWord(word) {
			continue
		}
		stemmed, err := snowball.Stem(word, "english", true)
		if err != nil {
			stemmed = word
		}

		if !seen[stemmed] {
			seen[stemmed] = true
			result = append(result, stemmed)
		}
	}
	return result
}
