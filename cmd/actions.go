package cmd

import (
	"fmt"
	"io/fs"
	// "os"

	"github.com/one2n-go-bootcamp/word-count/wc"
)

func Run(fs fs.FS, fileName string, lineCount, wordCount, charCount bool) (string, error) {
	var out string

	data, err := wc.ReadFile(fs, fileName)
	if err != nil {
		fmt.Println(err)
		return out, err
	}

	if lineCount {
		count := wc.CountLines(data)
		out = fmt.Sprintf("%s%8d ", out, count)
	}
	
	if wordCount {
		count := wc.CountWords(data)
		out = fmt.Sprintf("%s%8d ", out, count)
	}
	
	if charCount {
		count := wc.CountChars(data)
		out = fmt.Sprintf("%s%8d ", out, count)
	}

	out = fmt.Sprintf("%s%s", out, fileName)
	return out, nil
}