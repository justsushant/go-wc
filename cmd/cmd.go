package cmd

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"

	wc "github.com/one2n-go-bootcamp/go-wc/pkg"
)

// type to represent the user input
type WcInput struct {
	files     []string
	lineCount bool
	wordCount bool
	charCount bool
	stdin     io.Reader
	stdout    io.Writer
	stderr    io.Writer
}

func run(fSys fs.FS, input *WcInput) bool {
	// if no options provided
	if !input.lineCount && !input.wordCount && !input.charCount {
		input.lineCount = true
		input.wordCount = true
		input.charCount = true
	}

	option := []wc.WcOption{}
	for _, arg := range input.files {
		relPath, _ := getRelPath(fSys, arg)
		option = append(option, wc.WcOption{
			OrigPath:  arg,
			Path:      relPath,
			CountLine: input.lineCount,
			CountWord: input.wordCount,
			CountChar: input.charCount,
		})
	}
	if len(input.files) == 0 {
		option = append(option, wc.WcOption{
			Stdin:     input.stdin,
			CountLine: input.lineCount,
			CountWord: input.wordCount,
			CountChar: input.charCount,
		})
	}

	result := wc.Wc(fSys, option)

	return printResult(result, input)
}

func printResult(result []wc.WcResult, input *WcInput) bool {
	errFoundFlag := false
	for _, res := range result {
		if res.Err != nil {
			errFoundFlag = true
			fmt.Fprintln(input.stderr, res.Err.Error())
			continue
		}

		var output string
		if input.lineCount {
			output += fmt.Sprintf("%8d ", res.LineCount)
		}
		if input.wordCount {
			output += fmt.Sprintf("%8d ", res.WordCount)
		}
		if input.charCount {
			output += fmt.Sprintf("%8d ", res.CharCount)
		}
		output += res.Path
		fmt.Fprintln(input.stdout, output)
	}
	return errFoundFlag
}

func getRelPath(fSys fs.FS, path string) (relPath string, err error) {
	absPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	root := fmt.Sprintf("%s", fSys)
	relPath, err = filepath.Rel(root, absPath)
	if err != nil {
		return "", err
	}

	return relPath, nil
}
