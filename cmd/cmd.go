package cmd

import (
	"fmt"
	"io"
	"io/fs"

	wc "github.com/one2n-go-bootcamp/word-count/pkg"
)

func run(fSys fs.FS, args []string, lineCount, wordCount, charCount bool, stdout, stderr io.Writer) error {
	option := wc.WcOption{Path: args}

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

	printResult(option, result, stdout, stderr)
	return nil
}

func printResult(option wc.WcOption, result []wc.WcResult, stdout, stderr io.Writer) {
	for _, res := range result {
		if res.Err != nil {
			fmt.Fprintln(stderr, res.Err.Error())
			continue
		}

		var output string
		if option.CountLine {
			output += fmt.Sprintf("%8d ", res.LineCount)
		}
		if option.CountWord {
			output += fmt.Sprintf("%8d ", res.WordCount)
		}
		if option.CountChar {
			output += fmt.Sprintf("%8d ", res.CharCount)
		}
		output += res.Path
		fmt.Fprintln(stdout, output)
	}
}