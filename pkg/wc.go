package wc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"unicode"
)

var (
	ErrIsDirectory = errors.New("is a directory")
)

type WcOption struct {
	Path []string
	Stdin io.Reader
	CountLine bool
	CountWord bool
	CountChar bool
}

type WcResult struct {
	Path string
	LineCount int
	WordCount int
	CharCount int
	Err error
}

func Wc(fSys fs.FS, option WcOption) ([]WcResult, error) {
	result := []WcResult{}

	for _, path := range option.Path {
		r, cleanup, err := getReader(fSys, path, option.Stdin)
		if err != nil {
			result = append(result, WcResult{Err: err})
			continue
		}
		// defer cleanup()

		res, err := count(r, path, option)
		if err != nil {
			result = append(result, WcResult{Err: err})
			continue
		}
		result = append(result, res)
		cleanup()
	}

	if len(result) > 1 {
		result = append(result, calcTotal(result))
	}

	return result, nil
}

func count(r io.Reader, path string, option WcOption) (WcResult, error) {
	var lineCount, wordCount, charCount int
	spaceFlag := true

	var result WcResult
	result.Path = path

	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanBytes)
	for scanner.Scan() {
		// counting char
		charCount += 1

		// counting line
		if scanner.Text() == "\n" {
			lineCount += 1
		}

		// counting word
		// marks the current byte to be space, to be used in next iteration
		if unicode.IsSpace(rune(scanner.Bytes()[0])) {
			spaceFlag = true
		}
		// if previous byte was whitespace, and current one isn't, count it word
		if spaceFlag && !unicode.IsSpace(rune(scanner.Bytes()[0])) {
			wordCount += 1
			spaceFlag = false
		}
	}

	if err := scanner.Err(); err != nil {
		return WcResult{}, err
	}

	if option.CountLine {
		result.LineCount = lineCount
	}
	if option.CountWord {
		result.WordCount = wordCount
	}
	if option.CountChar {
		result.CharCount = charCount
	}

	return result, nil
}

func getReader(fSys fs.FS, path string, stdin io.Reader) (io.Reader, func(), error) {
	if path != "" {
		err := isValid(fSys, path)
		if err != nil {
			return nil, nil, err
		}

		file, err := fSys.Open(path)
		if err != nil {
			return nil, nil, err
		}
		return file, func() {file.Close()}, nil
	}

	return stdin, func() {}, nil
}

func isValid(fSys fs.FS, path string) error {
	fileInfo, err := fs.Stat(fSys, path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%s: %w", path, fs.ErrNotExist)
		}
		return fmt.Errorf("%s: %w", path, err)
	}

	// checks for directory
	if fileInfo.IsDir() {
		return fmt.Errorf("%s: %w", path, ErrIsDirectory)
	}

	// checks for permissions
	// looks hacky, might have to change later
	if fileInfo.Mode().Perm()&400 == 0 {
		return fmt.Errorf("%s: %w", path, fs.ErrPermission)
	}

	return nil
}

func calcTotal(result []WcResult) WcResult {
	total := WcResult{Path: "total"}

	for _, res := range result {
		total.LineCount += res.LineCount
		total.WordCount += res.WordCount
		total.CharCount += res.CharCount
	}

	return total
}
