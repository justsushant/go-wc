package cmd

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"

	wc "github.com/one2n-go-bootcamp/go-wc/pkg"
)

func run(fSys fs.FS, args []string, lineCount, wordCount, charCount bool, stdin io.Reader, stdout, stderr io.Writer) bool {
	// if no options provided
	if !lineCount == wordCount == charCount {
		lineCount = true
		wordCount = true
		charCount = true
	}

	option := []wc.WcOption{}
	for _, arg := range args {
		relPath, _ := getRelPath(fSys, arg)
		option = append(option, wc.WcOption{
			OrigPath:  arg,
			Path:      relPath,
			CountLine: lineCount,
			CountWord: wordCount,
			CountChar: charCount,
		})
	}
	if len(args) == 0 {
		option = append(option, wc.WcOption{
			Stdin:     stdin,
			CountLine: lineCount,
			CountWord: wordCount,
			CountChar: charCount,
		})
	}

	result := wc.Wc(fSys, option)

	return printResult(result, lineCount, wordCount, charCount, stdout, stderr)
}

func printResult(result []wc.WcResult, lineCount, wordCount, charCount bool, stdout, stderr io.Writer) bool {
	errFoundFlag := false
	for _, res := range result {
		if res.Err != nil {
			errFoundFlag = true
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
	return errFoundFlag
}

func getRelPath(fSys fs.FS, arg string) (relPath string, err error) {
	absPath, err := filepath.Abs(filepath.Clean(arg))
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	root := fmt.Sprintf("%s", fSys)
	// fmt.Println("Root: ", root)
	// fmt.Println("Abs Path: ", absPath)
	relPath, err = filepath.Rel(root, absPath)
	if err != nil {
		return "", err
	}

	return relPath, nil
}
