package wc

import (
	"bytes"
	"errors"
	// "fmt"
	"io/fs"
	"reflect"
	"testing"
	"testing/fstest"
)
func TestCount(t *testing.T) {
	testFS := fstest.MapFS{
		"file1.txt": {Data: []byte(""), Mode: 0755},
		"file2.txt": {Data: []byte("single_line"), Mode: 0755},
		"file3.txt": {Data: []byte("single line\nand\ndouble line\nin\nfile"), Mode: 0755},
		"file4.txt": {Data: []byte("\nI love mangoes,\tapples- but it applies to most fruits.\n??--ww"), Mode: 0755},
		"file5.txt": {Data: []byte("this file got permisson error"), Mode: 0000},
		"dir1": {Mode: fs.ModeDir},
	}

	testCases := []struct{
		name string
		path []string
		stdin []byte
		countLine bool
		countWord bool
		countChar bool
		result []WcResult
		expErr error
	}{
		{
			name: "wc -l with no matches",
			path: []string{"file1.txt"},
			countLine: true, 
			result: []WcResult{{Path: "file1.txt", LineCount: 0}}, 

		},
		{
			name: "wc -l with single match",
			path: []string{"file2.txt"},
			countLine: true, 
			result: []WcResult{{Path: "file2.txt", LineCount: 0}}, 

		},
		{
			name: "wc -l with multiple matches",
			path: []string{"file3.txt"},
			countLine: true, 
			result: []WcResult{{Path: "file3.txt", LineCount: 4}}, 

		},
		{
			name: "wc -w with no matches",
			path: []string{"file1.txt"},
			countWord: true, 
			result: []WcResult{{Path: "file1.txt", WordCount: 0}}, 

		},
		{
			name: "wc -w with single match",
			path: []string{"file2.txt"},
			countWord: true, 
			result: []WcResult{{Path: "file2.txt", WordCount: 1}}, 

		},
		{
			name: "wc -w with multiple matches",
			path: []string{"file3.txt"},
			countWord: true, 
			result: []WcResult{{Path: "file3.txt", WordCount: 7}}, 

		},
		{
			name: "wc -c with no matches",
			path: []string{"file1.txt"},
			countChar: true, 
			result: []WcResult{{Path: "file1.txt", CharCount: 0}}, 

		},
		{
			name: "wc -c with single match",
			path: []string{"file2.txt"},
			countChar: true, 
			result: []WcResult{{Path: "file2.txt", CharCount: 11}}, 

		},
		{
			name: "wc -c with multiple matches",
			path: []string{"file3.txt"},
			countChar: true, 
			result: []WcResult{{Path: "file3.txt", CharCount: 35}}, 

		},
		{
			name: "wc -lc with multiple matches",
			path: []string{"file3.txt"},
			countLine: true,
			countChar: true, 
			result: []WcResult{{Path: "file3.txt", LineCount: 4, CharCount: 35}}, 
		},
		{
			name: "wc -wc with multiple matches",
			path: []string{"file3.txt"},
			countWord: true,
			countChar: true, 
			result: []WcResult{{Path: "file3.txt", WordCount: 7, CharCount: 35}}, 
		},
		{
			name: "wc -lw with multiple matches",
			path: []string{"file3.txt"},
			countLine: true,
			countWord: true, 
			result: []WcResult{{Path: "file3.txt", LineCount: 4, WordCount: 7}}, 
		},
		{
			name: "wc -lwc with multiple matches",
			path: []string{"file3.txt"},
			countLine: true,
			countWord: true,
			countChar: true, 
			result: []WcResult{{Path: "file3.txt", LineCount: 4, WordCount: 7, CharCount: 35}}, 
		},
		{
			name: "wc -lwc with multiple files and multiple matches",
			path: []string{"file3.txt", "file2.txt"},
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
			name: "wc with stdin",
			stdin: []byte("this\nis\na\nfile"),
			countLine: true,
			countWord: true,
			countChar: true, 
			result: []WcResult{{LineCount: 3, WordCount: 4, CharCount: 14}}, 
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
			name: "wc over non-existent file",
			path: []string{"non-existent-file.txt"},
			expErr: fs.ErrNotExist,
			result: []WcResult{{Err: fs.ErrNotExist}},
		},
		{
			name: "wc over a file with permisson error",
			path: []string{"file5.txt"},
			expErr: fs.ErrPermission,
			result: []WcResult{{Err: fs.ErrPermission}},
		},
		{
			name: "wc over a directory",
			path: []string{"dir1"},
			expErr: ErrIsDirectory,
			result: []WcResult{{Err: ErrIsDirectory}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			want := tc.result
			option := []WcOption{}
			for _, p := range tc.path {
				option = append(option, WcOption{Path: p, CountLine: tc.countLine, CountWord: tc.countWord, CountChar: tc.countChar})
			}

			if tc.path == nil {
				option = append(option, WcOption{Stdin: bytes.NewReader(tc.stdin), CountLine: tc.countLine, CountWord: tc.countWord, CountChar: tc.countChar})
			}

			got, err := Wc(testFS, option)

			if tc.expErr != nil {
				// fmt.Println(err)

				// fmt.Println(want)
				// fmt.Println(got)
				
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

				t.Fatalf("Expected error %q but got %q", tc.expErr.Error(), err.Error())
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !reflect.DeepEqual(want, got) {
				t.Errorf("Expected %v but got %v", want, got)
			}
		})
	}
}