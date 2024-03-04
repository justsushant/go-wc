package cmd

import (
	"fmt"
	"io"
	"io/fs"

	wc "github.com/one2n-go-bootcamp/word-count/pkg"
)

func run(fSys fs.FS, args []string, lineCount, wordCount, charCount bool, stdin io.Reader, stdout, stderr io.Writer) error {
	// if no options provided
	if !lineCount == wordCount == charCount {
		lineCount = true
		wordCount = true
		charCount = true
	}

	option := []wc.WcOption{}
	for _, arg := range args {
		option = append(option, wc.WcOption{
			Path: arg, 
			CountLine: lineCount, 
			CountWord: wordCount, 
			CountChar: charCount,
		})
	}
	if len(args) == 0 {
		option = append(option, wc.WcOption{
			Stdin: stdin,
			CountLine: lineCount, 
			CountWord: wordCount, 
			CountChar: charCount,
		})
	}

	result, err := wc.Wc(fSys, option)
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		return err
	}

	printResult(result, lineCount, wordCount, charCount, stdout, stderr)
	return nil
}

func printResult(result []wc.WcResult, lineCount, wordCount, charCount bool, stdout, stderr io.Writer) {
	for _, res := range result {
		if res.Err != nil {
			fmt.Fprintln(stderr, res.Err.Error())
			continue
		}

		var output string
		if lineCount {
			output += fmt.Sprintf("%8d ", res.LineCount)
		}
		if wordCount {
			output += fmt.Sprintf("%8d ", res.WordCount)
		}
		if charCount {
			output += fmt.Sprintf("%8d ", res.CharCount)
		}
		output += res.Path
		fmt.Fprintln(stdout, output)
	}
}
