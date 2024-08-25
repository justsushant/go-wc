package wc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"sync"
	"unicode"
)

var (
	ErrIsDirectory = errors.New("is a directory")
)

type WcOption struct {
	OrigPath  string
	Path      string
	Stdin     io.Reader
	CountLine bool
	CountWord bool
	CountChar bool
}

type WcResult struct {
	Path      string
	LineCount int
	WordCount int
	CharCount int
	Err       error
}

func Wc(fSys fs.FS, option []WcOption) []WcResult {
	var wg sync.WaitGroup
	var outputChans = make([]chan WcResult, len(option)) // to aggregate the channels
	result := []WcResult{}

	for i, op := range option {
		outputChan := make(chan WcResult, len(option))
		outputChans[i] = outputChan

		// launches a new go routine for each file
		wg.Add(1)
		go func(fSys fs.FS, op WcOption, outputChan chan WcResult) {
			defer wg.Done()
			defer close(outputChan)

			// getting the reader and making checks
			r, cleanup, err := getReader(fSys, op.Path, op.Stdin)
			if err != nil {
				outputChan <- WcResult{Err: err}
				return
			}
			defer cleanup()

			// counting operation
			result, err := count(r, op)
			if err != nil {
				outputChan <- WcResult{Err: err}
				return
			}
			outputChan <- result
		}(fSys, op, outputChan)
	}
	wg.Wait()

	// collates data from all the channels
	for _, outputChan := range outputChans {
		result = append(result, <-outputChan)
	}

	// adds the total if more than one file
	if len(result) > 1 {
		result = append(result, calcTotal(result))
	}

	return result
}

func count(r io.Reader, option WcOption) (WcResult, error) {
	var lineCount, wordCount, charCount int
	spaceFlag := true // to keep track of previous whitespace

	var result WcResult
	result.Path = option.OrigPath

	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanBytes)
	for scanner.Scan() {
		// counting char
		if option.CountChar {
			charCount++
		}

		// counting line
		if option.CountLine {
			if scanner.Text() == "\n" {
				lineCount++
			}
		}

		// counting word
		if option.CountWord {
			// marks the current byte to be space, to be used in next iteration
			if unicode.IsSpace(rune(scanner.Bytes()[0])) {
				spaceFlag = true
			}
			// if previous byte was whitespace, and current one isn't, count it word
			if spaceFlag && !unicode.IsSpace(rune(scanner.Bytes()[0])) {
				wordCount++
				spaceFlag = false
			}
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
		return file, func() { file.Close() }, nil
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
	// possible alternative: fileInfo.Mode().Perm()&(1<<8) == 0
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
