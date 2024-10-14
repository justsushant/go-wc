package wc

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"reflect"
	"testing"
	"testing/fstest"
	"testing/iotest"
)

func TestWc(t *testing.T) {
	testFS := fstest.MapFS{
		"file1.txt":  {Data: []byte(""), Mode: 0755},
		"file2.txt":  {Data: []byte("single_line"), Mode: 0755},
		"file3.txt":  {Data: []byte("single line\nand\ndouble line\nin\nfile"), Mode: 0755},
		"file4.txt":  {Data: []byte("\nI love mangoes,\tapples- but it applies to most fruits.\n??--ww"), Mode: 0755},
		"file5.txt":  {Data: []byte("this file got permisson error"), Mode: 0000},
		"file6.txt":  {Data: []byte("dummy file 6"), Mode: 0755},
		"file7.tar":  {Data: []byte("dummy file 7"), Mode: 0755},
		"file8.txt":  {Data: []byte(""), Mode: 0755},
		"file9.jpg":  {Data: []byte("dummy file 9"), Mode: 0755},
		"file10.mov": {Data: []byte("dummy file 10"), Mode: 0755},
		"file11.png": {Data: []byte("dummy file 11"), Mode: 0755},
		"file12.md":  {Data: []byte("dummy file 11"), Mode: 0755},
		"dir1":       {Mode: fs.ModeDir},
	}

	testCases := []struct {
		name       string
		path       []string
		stdin      []byte
		countLine  bool
		countWord  bool
		countChar  bool
		excludeExt []string
		includeExt []string
		result     []WcResult
		expErr     error
	}{
		{
			name:      "wc -l with no matches",
			path:      []string{"file1.txt"},
			countLine: true,
			result:    []WcResult{{Path: "file1.txt", LineCount: 0}},
		},
		{
			name:      "wc -l with single match",
			path:      []string{"file2.txt"},
			countLine: true,
			result:    []WcResult{{Path: "file2.txt", LineCount: 0}},
		},
		{
			name:      "wc -l with multiple matches",
			path:      []string{"file3.txt"},
			countLine: true,
			result:    []WcResult{{Path: "file3.txt", LineCount: 4}},
		},
		{
			name:      "wc -w with no matches",
			path:      []string{"file1.txt"},
			countWord: true,
			result:    []WcResult{{Path: "file1.txt", WordCount: 0}},
		},
		{
			name:      "wc -w with single match",
			path:      []string{"file2.txt"},
			countWord: true,
			result:    []WcResult{{Path: "file2.txt", WordCount: 1}},
		},
		{
			name:      "wc -w with multiple matches",
			path:      []string{"file3.txt"},
			countWord: true,
			result:    []WcResult{{Path: "file3.txt", WordCount: 7}},
		},
		{
			name:      "wc -c with no matches",
			path:      []string{"file1.txt"},
			countChar: true,
			result:    []WcResult{{Path: "file1.txt", CharCount: 0}},
		},
		{
			name:      "wc -c with single match",
			path:      []string{"file2.txt"},
			countChar: true,
			result:    []WcResult{{Path: "file2.txt", CharCount: 11}},
		},
		{
			name:      "wc -c with multiple matches",
			path:      []string{"file3.txt"},
			countChar: true,
			result:    []WcResult{{Path: "file3.txt", CharCount: 35}},
		},
		{
			name:      "wc -lc with multiple matches",
			path:      []string{"file3.txt"},
			countLine: true,
			countChar: true,
			result:    []WcResult{{Path: "file3.txt", LineCount: 4, CharCount: 35}},
		},
		{
			name:      "wc -wc with multiple matches",
			path:      []string{"file3.txt"},
			countWord: true,
			countChar: true,
			result:    []WcResult{{Path: "file3.txt", WordCount: 7, CharCount: 35}},
		},
		{
			name:      "wc -lw with multiple matches",
			path:      []string{"file3.txt"},
			countLine: true,
			countWord: true,
			result:    []WcResult{{Path: "file3.txt", LineCount: 4, WordCount: 7}},
		},
		{
			name:      "wc -lwc with multiple matches",
			path:      []string{"file3.txt"},
			countLine: true,
			countWord: true,
			countChar: true,
			result:    []WcResult{{Path: "file3.txt", LineCount: 4, WordCount: 7, CharCount: 35}},
		},
		{
			name:      "wc -lwc with multiple files and multiple matches",
			path:      []string{"file3.txt", "file2.txt"},
			countLine: true,
			countWord: true,
			countChar: true,
			result: []WcResult{
				{Path: "file3.txt", LineCount: 4, WordCount: 7, CharCount: 35},
				{Path: "file2.txt", LineCount: 0, WordCount: 1, CharCount: 11},
				{Path: "total", LineCount: 4, WordCount: 8, CharCount: 46},
			},
		},
		{
			name:      "wc with stdin",
			stdin:     []byte("this\nis\na\nfile"),
			countLine: true,
			countWord: true,
			countChar: true,
			result:    []WcResult{{LineCount: 3, WordCount: 4, CharCount: 14}},
		},
		{
			name:      "wc with empty file",
			path:      []string{"file8.txt"},
			countLine: true,
			countWord: true,
			countChar: true,
			result:    []WcResult{{Path: "file8.txt", LineCount: 0, WordCount: 0, CharCount: 0}},
		},
		{
			name:      "wc with one empty file and one valid file with text",
			path:      []string{"file3.txt", "file8.txt"},
			countLine: true,
			countWord: true,
			countChar: true,
			result: []WcResult{
				{Path: "file3.txt", LineCount: 4, WordCount: 7, CharCount: 35},
				{Path: "file8.txt", LineCount: 0, WordCount: 0, CharCount: 0},
				{Path: "total", LineCount: 4, WordCount: 7, CharCount: 35},
			},
		},
		// {
		// 	name:      "wc -lwc with multiple matches (random symbols and spaces)",
		// 	path:      []string{"file4.txt"},
		// 	countLine: true,
		// 	countWord: true,
		// 	countChar: true,
		// 	result:    []WcResult{{Path: "file4.txt", LineCount: 2, WordCount: 11, CharCount: 64}},
		// },
		{
			name:   "wc over non-existent file",
			path:   []string{"non-existent-file.txt"},
			expErr: fs.ErrNotExist,
			result: []WcResult{{Err: fs.ErrNotExist}},
		},
		{
			name:   "wc over a file with permisson error",
			path:   []string{"file5.txt"},
			expErr: fs.ErrPermission,
			result: []WcResult{{Err: fs.ErrPermission}},
		},
		{
			name:   "wc over a directory",
			path:   []string{"dir1"},
			expErr: ErrIsDirectory,
			result: []WcResult{{Err: ErrIsDirectory}},
		},
		{
			name:       "wc with exclude ext flag with one exclude ext",
			path:       []string{"file7.tar"},
			countLine:  true,
			countWord:  true,
			countChar:  true,
			excludeExt: []string{"tar"},
			result:     []WcResult{},
		},
		{
			name:       "wc with exclude ext flag with one valid file and one exclude ext",
			path:       []string{"file3.txt", "file7.tar"},
			countLine:  true,
			countWord:  true,
			countChar:  true,
			excludeExt: []string{"tar"},
			result:     []WcResult{{Path: "file3.txt", LineCount: 4, WordCount: 7, CharCount: 35}},
		},
		{
			name:       "wc with exclude ext flag with two valid file and one exclude ext",
			path:       []string{"file3.txt", "file7.tar", "file2.txt"},
			countLine:  true,
			countWord:  true,
			countChar:  true,
			excludeExt: []string{"tar"},
			result: []WcResult{
				{Path: "file3.txt", LineCount: 4, WordCount: 7, CharCount: 35},
				{Path: "file2.txt", LineCount: 0, WordCount: 1, CharCount: 11},
				{Path: "total", LineCount: 4, WordCount: 8, CharCount: 46},
			},
		},
		{
			name:       "wc with exclude ext flag with one valid file and three exclude ext",
			path:       []string{"file3.txt", "file9.jpg", "file10.mov", "file11.png"},
			countLine:  true,
			countWord:  true,
			countChar:  true,
			excludeExt: []string{"jpg", "mov", "png"},
			result: []WcResult{
				{Path: "file3.txt", LineCount: 4, WordCount: 7, CharCount: 35},
			},
		},
		{
			name:       "wc with include ext flag with one include ext",
			path:       []string{"file3.txt"},
			countLine:  true,
			countWord:  true,
			countChar:  true,
			includeExt: []string{"txt"},
			result:     []WcResult{{Path: "file3.txt", LineCount: 4, WordCount: 7, CharCount: 35}},
		},
		{
			name:       "wc with include ext flag with one other ext and one include ext",
			path:       []string{"file3.txt", "file7.tar"},
			countLine:  true,
			countWord:  true,
			countChar:  true,
			includeExt: []string{"txt"},
			result:     []WcResult{{Path: "file3.txt", LineCount: 4, WordCount: 7, CharCount: 35}},
		},
		{
			name:       "wc with include ext flag with two other ext and one include ext",
			path:       []string{"file3.txt", "file7.tar", "file2.txt"},
			countLine:  true,
			countWord:  true,
			countChar:  true,
			includeExt: []string{"tar"},
			result: []WcResult{
				{Path: "file7.tar", LineCount: 0, WordCount: 3, CharCount: 12},
			},
		},
		{
			name:       "wc with exclude ext flag with two valid ext and three other ext",
			path:       []string{"file3.txt", "file9.jpg", "file10.mov", "file11.png", "file12.md"},
			countLine:  true,
			countWord:  true,
			countChar:  true,
			includeExt: []string{"txt", "md"},
			result: []WcResult{
				{Path: "file3.txt", LineCount: 4, WordCount: 7, CharCount: 35},
				{Path: "file12.md", LineCount: 0, WordCount: 3, CharCount: 13},
				{Path: "total", LineCount: 4, WordCount: 10, CharCount: 48},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			want := tc.result
			option := []WcOption{}
			for _, p := range tc.path {
				option = append(option, WcOption{
					OrigPath:   p,
					Path:       p,
					CountLine:  tc.countLine,
					CountWord:  tc.countWord,
					CountChar:  tc.countChar,
					ExcludeExt: tc.excludeExt,
					IncludeExt: tc.includeExt,
				})
			}

			if tc.path == nil {
				option = append(option, WcOption{
					Stdin:     bytes.NewReader(tc.stdin),
					CountLine: tc.countLine,
					CountWord: tc.countWord,
					CountChar: tc.countChar,
				})
			}

			got := Wc(testFS, option)

			// iterating over the WcResult slice
			// we're using boolean flags to figure out if any errors were present in any of the result
			for _, g := range got {
				if tc.expErr != nil {
					isErrNilFlag := false
					for _, res := range got {
						if res.Err != nil {
							isErrNilFlag = true
							break
						}
					}
					if !isErrNilFlag {
						t.Errorf("Expected error %q but got nil", tc.expErr.Error())
					}

					errFoundFlag := false
					for _, res := range got {
						if errors.Is(res.Err, tc.expErr) {
							errFoundFlag = true
							break
						}
					}
					if errFoundFlag {
						return
					}

					t.Fatalf("Expected error %q but got %q", tc.expErr.Error(), g.Err.Error())
				}

				if g.Err != nil {
					t.Fatalf("Unexpected error: %q", g.Err.Error())
				}

				// this warning has to be fixed since can't use it with the errors
				if !reflect.DeepEqual(want, got) {
					t.Errorf("Expected %v but got %v", want, got)
				}
			}
		})
	}
}

func TestCount(t *testing.T) {
	ErrForTesting := errors.New("error for testing")

	tt := []struct {
		name   string
		reader io.Reader
		option WcOption
		expOut WcResult
		expErr error
	}{
		{
			name:   "normal happy case with three options",
			reader: bytes.NewReader([]byte("single line\nand\ndouble line\nin\nfile")),
			option: WcOption{
				CountLine: true,
				CountWord: true,
				CountChar: true,
				OrigPath:  "/xyz",
			},
			expOut: WcResult{
				Path:      "/xyz",
				LineCount: 4,
				WordCount: 7,
				CharCount: 35,
				Err:       nil,
			},
			expErr: nil,
		},
		{
			name:   "normal happy case with two options",
			reader: bytes.NewReader([]byte("single line\nand\ndouble line\nin\nfile")),
			option: WcOption{
				CountLine: true,
				CountWord: true,
				OrigPath:  "/xyz",
			},
			expOut: WcResult{
				Path:      "/xyz",
				LineCount: 4,
				WordCount: 7,
				Err:       nil,
			},
			expErr: nil,
		},
		{
			name:   "error unhappy case with error prone reader",
			reader: iotest.ErrReader(ErrForTesting),
			option: WcOption{
				CountLine: true,
				CountWord: true,
				OrigPath:  "/xyz",
			},
			expOut: WcResult{},
			expErr: ErrForTesting,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			result, err := count(tc.reader, tc.option)

			if tc.expErr != nil {
				if err == nil {
					t.Fatalf("Expected error %q but got nil", tc.expErr.Error())
				}

				if !errors.Is(err, tc.expErr) {
					t.Fatalf("Expected error %q but got %v", tc.expErr.Error(), err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %q", err.Error())
			}

			// avoiding the usage of reflect.DeepEqual because not a good practice to use it with errors
			if tc.expOut.Path != result.Path {
				t.Errorf("Expected the path %v but got %v", tc.expOut.Path, result.Path)
			}
			if tc.expOut.LineCount != result.LineCount {
				t.Errorf("Expected the line count %v but got %v", tc.expOut.LineCount, result.LineCount)
			}
			if tc.expOut.WordCount != result.WordCount {
				t.Errorf("Expected the word count %v but got %v", tc.expOut.WordCount, result.WordCount)
			}
			if tc.expOut.CharCount != result.CharCount {
				t.Errorf("Expected the char count %v but got %v", tc.expOut.CharCount, result.CharCount)
			}
			if !errors.Is(result.Err, tc.expOut.Err) {
				t.Errorf("Expected the error %v but got %v", tc.expOut.Err, result.Err)
			}
		})
	}
}

func TestGetReader(t *testing.T) {
	testFS := fstest.MapFS{
		"file1.txt": {Data: []byte(""), Mode: 0755},
		"file2.txt": {Data: []byte("single_line"), Mode: 0755},
		"file3.txt": {Data: []byte("single line\nand\ndouble line\nin\nfile"), Mode: 0755},
		"file4.txt": {Data: []byte("\nI love mangoes,\tapples- but it applies to most fruits.\n??--ww"), Mode: 0755},
		"file5.txt": {Data: []byte("this file got permisson error"), Mode: 0000},
		"dir1":      {Mode: fs.ModeDir},
	}

	tt := []struct {
		name            string
		option          WcOption
		expTextInReader string
		expErr          error
	}{
		{
			name: "nomal happy case with valid path",
			option: WcOption{
				Path:  "file2.txt",
				Stdin: nil,
			},
			expTextInReader: "single_line",
			expErr:          nil,
		},
		{
			name: "nomal happy case with empty path and valid stdin",
			option: WcOption{
				Path:  "",
				Stdin: bytes.NewReader([]byte("text from stdin\n here")),
			},
			expTextInReader: "text from stdin\n here",
			expErr:          nil,
		},
		{
			name: "error unhappy case with invalid path of file",
			option: WcOption{
				Path:  "invalid_file.txt",
				Stdin: nil,
			},
			expTextInReader: "",
			expErr:          fs.ErrNotExist,
		},
		{
			name: "error unhappy case with valid path of directory",
			option: WcOption{
				Path:  "dir1",
				Stdin: nil,
			},
			expTextInReader: "",
			expErr:          ErrIsDirectory,
		},
		{
			name: "error unhappy case with valid path of file but permisson error",
			option: WcOption{
				Path:  "file5.txt",
				Stdin: nil,
			},
			expTextInReader: "",
			expErr:          fs.ErrPermission,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			reader, cleanup, err := getReader(testFS, tc.option)
			if tc.expErr != nil {
				if err == nil {
					t.Fatalf("Expected error %q but got nil", tc.expErr.Error())
				}

				if !errors.Is(err, tc.expErr) {
					t.Fatalf("Expected error %q but got %q", tc.expErr.Error(), err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %q", err.Error())
			}

			// checking if the reader has the same text as we expected
			data, err := io.ReadAll(reader)
			if err != nil {
				t.Fatalf("Unexpected error while reading from io.Reader, %q", err.Error())
			}
			if string(data) != tc.expTextInReader {
				t.Errorf("Expected the reader to have %q but got %q", tc.expTextInReader, string(data))
			}

			// TODO: testing the behaviour of cleanup function
			// maybe we have to find a way to check if the file is still open (that would be too coupled to the implementation)
			// we can also do some kind of mock counter thing where the cleanup function would be injected from arguments and we will check if the counter has been incremented
			// very unclear what to do here
			cleanup()
		})
	}
}

func TestIsValid(t *testing.T) {
	testFS := fstest.MapFS{
		"file1.txt": {Data: []byte(""), Mode: 0755},
		"file2.txt": {Data: []byte("single_line"), Mode: 0755},
		"file3.txt": {Data: []byte("single line\nand\ndouble line\nin\nfile"), Mode: 0755},
		"file4.txt": {Data: []byte("\nI love mangoes,\tapples- but it applies to most fruits.\n??--ww"), Mode: 0755},
		"file5.txt": {Data: []byte("this file got permisson error"), Mode: 0000},
		"file6.txt": {Data: []byte("dummy file 6"), Mode: 0755},
		"dir1":      {Mode: fs.ModeDir},
	}

	tt := []struct {
		name            string
		option          WcOption
		expOut          bool
		expTextInReader string
		expErr          error
	}{
		{
			name:   "nomal happy case with valid path",
			expOut: true,
			option: WcOption{
				Path: "file2.txt",
			},
			expErr: nil,
		},
		{
			name:   "error unhappy case with empty path",
			expOut: false,
			option: WcOption{
				Path: "",
			},
			expErr: fs.ErrNotExist,
		},
		{
			name: "error unhappy case with invalid path of file",
			option: WcOption{
				Path: "xyz.txt",
			},
			expOut: false,
			expErr: fs.ErrNotExist,
		},
		{
			name: "error unhappy case with valid path of directory",
			option: WcOption{
				Path: "dir1",
			},
			expOut: false,
			expErr: ErrIsDirectory,
		},
		{
			name: "error unhappy case with valid path of file but permisson error",
			option: WcOption{
				Path: "file5.txt",
			},
			expOut: false,
			expErr: fs.ErrPermission,
		},
		{
			name: "exclude file extension",
			option: WcOption{
				Path:       "file6.txt",
				ExcludeExt: []string{"txt"},
			},
			expOut: false,
			expErr: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			gotOk, gotErr := isValid(testFS, tc.option)

			if tc.expErr != nil {
				if gotErr == nil {
					t.Fatalf("Expected error %q but got nil", tc.expErr.Error())
				}

				if !errors.Is(gotErr, tc.expErr) {
					t.Fatalf("Expected error %q but got %q", tc.expErr.Error(), gotErr.Error())
				}

				return
			}

			if gotErr != nil {
				t.Fatalf("Unexpected error: %q", gotErr.Error())
			}

			if gotOk != tc.expOut {
				t.Errorf("Expected %v but got %v", tc.expOut, gotErr)
			}

		})
	}
}

func TestCalcTotal(t *testing.T) {
	tt := []struct {
		name   string
		input  []WcResult
		expOut WcResult
	}{
		{
			name: "nomal case with all non-zero fields",
			input: []WcResult{
				{
					LineCount: 1,
					WordCount: 2,
					CharCount: 3,
				},
				{
					LineCount: 4,
					WordCount: 5,
					CharCount: 6,
				},
				{
					LineCount: 7,
					WordCount: 8,
					CharCount: 9,
				},
			},
			expOut: WcResult{
				Path:      "total",
				LineCount: 12,
				WordCount: 15,
				CharCount: 18,
			},
		},
		{
			name: "nomal case with all zero fields",
			input: []WcResult{
				{
					LineCount: 0,
					WordCount: 0,
					CharCount: 0,
				},
				{
					LineCount: 0,
					WordCount: 0,
					CharCount: 0,
				},
			},
			expOut: WcResult{
				Path:      "total",
				LineCount: 0,
				WordCount: 0,
				CharCount: 0,
			},
		},
		{
			name:  "error case with no result elements",
			input: []WcResult{},
			expOut: WcResult{
				Path: "total",
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := calcTotal(tc.input)

			// avoiding the usage of reflect.DeepEqual because not a good practice to use it with errors
			if tc.expOut.Path != got.Path {
				t.Errorf("Expected the path %v but got %v", tc.expOut.Path, got.Path)
			}
			if tc.expOut.LineCount != got.LineCount {
				t.Errorf("Expected the line count %v but got %v", tc.expOut.LineCount, got.LineCount)
			}
			if tc.expOut.WordCount != got.WordCount {
				t.Errorf("Expected the word count %v but got %v", tc.expOut.WordCount, got.WordCount)
			}
			if tc.expOut.CharCount != got.CharCount {
				t.Errorf("Expected the char count %v but got %v", tc.expOut.CharCount, got.CharCount)
			}
		})
	}
}
