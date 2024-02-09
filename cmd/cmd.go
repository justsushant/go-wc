package cmd

import (
	"fmt"
	"io"
	"io/fs"

	wc "github.com/one2n-go-bootcamp/word-count/pkg"
)

func run(fSys fs.FS, args []string, lineCount, wordCount, charCount bool, stdout, stderr io.Writer) error {
	option := wc.WcOption{Path:args[0]}

	// if no options provided
	if !lineCount == wordCount == charCount {
		option.CountLine = true
		option.CountWord = true
		option.CountChar = true
	} else {
		option.CountLine = lineCount
		option.CountWord = wordCount
		option.CountChar = charCount
	}

	result, err := wc.Wc(fSys, option)
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		return err
	}

	output := formatResult(option, result)
	fmt.Fprintln(stdout, output)
	return nil
}

func formatResult(option wc.WcOption, result wc.WcResult) string {
	var output string

	if option.CountLine {
		output += fmt.Sprintf("%8d ", result.LineCount)
	}
	if option.CountWord {
		output += fmt.Sprintf("%8d ", result.WordCount)
	}
	if option.CountChar {
		output += fmt.Sprintf("%8d ", result.CharCount)
	}

	output += option.Path
	return output
}