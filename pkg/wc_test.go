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
		"file1.txt": {Data: []byte(""), Mode: 0755},
		"file2.txt": {Data: []byte("single_line"), Mode: 0755},
		"file3.txt": {Data: []byte("single line\nand\ndouble line\nin\nfile"), Mode: 0755},
		"file4.txt": {Data: []byte("\nI love mangoes,\tapples- but it applies to most fruits.\n??--ww"), Mode: 0755},
		"file5.txt": {Data: []byte("this file got permisson error"), Mode: 0000},
		"dir1":      {Mode: fs.ModeDir},
	}

	testCases := []struct {
		name      string
		path      []string
		stdin     []byte
		countLine bool
		countWord bool
		countChar bool
		result    []WcResult
		expErr    error
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
		// {
		// 	name: "wc -lwc with multiple matches (random symbols and spaces)",
		// 	path: []string{"file4.txt"},
		// 	countLine: true,
		// 	countWord: true,
		// 	countChar: true,
		// 	result: []WcResult{{Path: "file4.txt", LineCount: 2, WordCount: 11, CharCount: 64}},
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			want := tc.result
			option := []WcOption{}
			for _, p := range tc.path {
				option = append(option, WcOption{OrigPath: p, Path: p, CountLine: tc.countLine, CountWord: tc.countWord, CountChar: tc.countChar})
			}

			if tc.path == nil {
				option = append(option, WcOption{Stdin: bytes.NewReader(tc.stdin), CountLine: tc.countLine, CountWord: tc.countWord, CountChar: tc.countChar})
			}

			got := Wc(testFS, option)

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
					t.Fatalf("Unexpected error: %v", g.Err)
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
					t.Fatalf("Expected error %v but got nil", tc.expErr)
				}

				if !errors.Is(err, tc.expErr) {
					t.Fatalf("Expected error %v but got %v", tc.expErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
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
