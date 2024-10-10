package wc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"slices"
	"sync"
	"unicode"
)

var (
	ErrIsDirectory = errors.New("is a directory")
)

const MAX_OPEN_FILE_DESCRIPTORS = 1024

type WcOption struct {
	OrigPath   string
	Path       string
	Stdin      io.Reader
	CountLine  bool
	CountWord  bool
	CountChar  bool
	ExcludeExt []string
	IncludeExt []string
}

type WcResult struct {
	Path      string
	LineCount int
	WordCount int
	CharCount int
	Err       error
}

func Wc(fSys fs.FS, option []WcOption) []WcResult {
	var openFileLimit int = MAX_OPEN_FILE_DESCRIPTORS
	cond := sync.NewCond(&sync.Mutex{})

	var wg sync.WaitGroup
	var outputChans = make([]chan *WcResult, len(option)) // to aggregate the channels
	result := []WcResult{}

	for i, op := range option {
		outputChan := make(chan *WcResult, len(option))
		outputChans[i] = outputChan

		// launches a new go routine for each file
		wg.Add(1)
		go func(fSys fs.FS, op WcOption, outputChan chan *WcResult, cond *sync.Cond) {
			defer wg.Done()
			defer close(outputChan)

			// decrement the openFileLimit since we're opening the file
			// lock the mutex before decrementing the openFileLimit
			// check if openFileLimit > 0; if its not, means we're at the limit
			// go routine will wait till the limit is resolved
			// unlock after decrementing the openFileLimit
			cond.L.Lock()
			for openFileLimit <= 0 {
				cond.Wait()
			}
			openFileLimit--
			cond.L.Unlock()

			// getting the reader and making checks
			r, cleanup, err := getReader(fSys, op)

			defer func() {
				// to free file resources
				cleanup()

				// increment the openFileLimit since we're closing the file
				// lock the mutex before incrementing the openFileLimit
				// signal other gouroutine to start
				// unlock after incrementing the openFileLimit
				cond.L.Lock()
				openFileLimit++
				// maybe we can broadcast here, since in an extreme edge case,
				// all gouroutines could hang on waiting indefinetly because upon signal, those goroutines still didnt fulfilled the condition
				cond.Signal()
				cond.L.Unlock()
			}()

			if err != nil {
				outputChan <- &WcResult{Err: err}
				return
			}

			// if reader is nil, then return
			// means there was no error, but no reader either
			// for eg, in case of exclude extension
			if r == nil {
				return
			}

			// counting operation
			result, err := count(r, op)
			if err != nil {
				outputChan <- &WcResult{Err: err}
				return
			}
			outputChan <- &result
		}(fSys, op, outputChan, cond)
	}
	wg.Wait()

	// collates data from all the channels
	for _, outputChan := range outputChans {
		output := <-outputChan

		if output == nil {
			continue
		}
		result = append(result, *output)
	}

	// adds the total if more than one file
	if len(result) > 1 {
		total := calcTotal(result)
		result = append(result, total)
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

func getReader(fSys fs.FS, option WcOption) (io.Reader, func(), error) {
	if option.Path != "" {
		ok, err := isValid(fSys, option)
		if err != nil {
			return nil, func() {}, err
		}
		if !ok {
			return nil, func() {}, nil
		}

		file, err := fSys.Open(option.Path)
		if err != nil {
			return nil, nil, err
		}
		return file, func() { file.Close() }, nil
	}

	return option.Stdin, func() {}, nil
}

func isValid(fSys fs.FS, option WcOption) (bool, error) {
	fileInfo, err := fs.Stat(fSys, option.Path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, fmt.Errorf("%s: %w", option.Path, fs.ErrNotExist)
		}
		return false, fmt.Errorf("%s: %w", option.Path, err)
	}

	// checks for directory
	if fileInfo.IsDir() {
		return false, fmt.Errorf("%s: %w", option.Path, ErrIsDirectory)
	}

	// checks for permissions
	// looks hacky, might have to change later
	// possible alternative: fileInfo.Mode().Perm()&(1<<8) == 0
	if fileInfo.Mode().Perm()&400 == 0 {
		return false, fmt.Errorf("%s: %w", option.Path, fs.ErrPermission)
	}

	// check if extension to be exlcuded
	// slicing to remove the dot (.) from start
	ext := filepath.Ext(fileInfo.Name())[1:]

	// if extension matches with exclude extension flag, don't count it
	if slices.Contains(option.ExcludeExt, ext) {
		return false, nil
	}

	// if include extension flag was passed, only count the approved extension
	if option.IncludeExt != nil {
		if slices.Contains(option.IncludeExt, ext) {
			return true, nil
		}
		return false, nil
	}

	return true, nil
}

func calcTotal(result []WcResult) WcResult {
	total := &WcResult{Path: "total"}

	for _, res := range result {
		total.LineCount += res.LineCount
		total.WordCount += res.WordCount
		total.CharCount += res.CharCount
	}

	return *total
}
