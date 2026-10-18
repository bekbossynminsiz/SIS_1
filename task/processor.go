package task

import (
	"strings"
	"unicode/utf8"
)


type Processor interface {
	Process(text string) (Result, error)
}

type TextAnalyzer struct{}

func (TextAnalyzer) Process(text string) (Result, error) {
	words := strings.Fields(text) 

	longest := ""
	for _, w := range words {
		if utf8.RuneCountInString(w) > utf8.RuneCountInString(longest) {
			longest = w
		}
	}

	return Result{
		Words:       len(words),
		Characters:  utf8.RuneCountInString(text),
		LongestWord: longest,
	}, nil
}
