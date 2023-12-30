package cmd

import (
	"fmt"
	"io/fs"

	"github.com/one2n-go-bootcamp/word-count/wc"
)

func Run(fs fs.FS, fileName string, lineCount, wordCount, charCount bool) (string, error) {
	var options []wc.Option

	data, err := wc.ReadFile(fs, fileName)
	if err != nil {
		return "", err
	}

	if !lineCount && !wordCount && !charCount {
		options = append(options, wc.CountLines, wc.CountWords, wc.CountChars)
	}

	if lineCount {
		options = append(options, wc.CountLines)
	}
	
	if wordCount {
		options = append(options, wc.CountWords)
	}
	
	if charCount {
		options = append(options, wc.CountChars)
	}


	out := Count(data, options...)
	out += fileName + "\n"
	return out, nil
}

// strings.builder can be used here
func Count(content []byte, options ...wc.Option) string {
	var out string
	for _, option := range options {
		out += fmt.Sprintf("%8d ", option(content))
	}

	return out
}