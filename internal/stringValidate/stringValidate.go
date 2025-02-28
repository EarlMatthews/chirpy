package stringValidate
import (
	"strings"
)

func ReplaceWord(input string, targets []string) string {
	words := strings.Fields(input)
	for i, word := range words {
		for _, target := range targets {
			if strings.EqualFold(word, target) {
				words[i] = "****"
				break // Exit inner loop once a match is found
			}
		}
	}
	return strings.Join(words, " ")
}