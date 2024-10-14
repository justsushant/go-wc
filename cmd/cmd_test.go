package cmd

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"

	wc "github.com/one2n-go-bootcamp/go-wc/pkg"
)

func TestRun(t *testing.T) {
	testCases := []struct {
		name       string
		path       []string
		stdin      []byte
		countLine  bool
		countWord  bool
		countChar  bool
		excludeExt []string
		includeExt []string
		expResult  string
		expError   error
	}{
		{
			name:     "wc over non-existent-file",
			path:     []string{"../testdata/cmd_test/non-existent-file.txt"},
			expError: fs.ErrNotExist,
		},
		{
			name:     "wc over unsufficient permission file",
			path:     []string{"../testdata/cmd_test/file5.txt"},
			expError: fs.ErrPermission,
		},
		{
			name:     "wc over directory",
			path:     []string{"../testdata/cmd_test/testdir"},
			expError: wc.ErrIsDirectory,
		},
		{
			name:      "wc -l with no matches",
			path:      []string{"../testdata/cmd_test/file2.txt"},
			countLine: true,
			expResult: "       0 ../testdata/cmd_test/file2.txt\n",
		},
		{
			name:      "wc -l with matches",
			path:      []string{"../testdata/cmd_test/file1.txt"},
			countLine: true,
			expResult: "       5 ../testdata/cmd_test/file1.txt\n",
		},
		{
			name:      "wc -w with no matches",
			path:      []string{"../testdata/cmd_test/file2.txt"},
			countWord: true,
			expResult: "       0 ../testdata/cmd_test/file2.txt\n",
		},
		{
			name:      "wc -w with matches",
			path:      []string{"../testdata/cmd_test/file1.txt"},
			countWord: true,
			expResult: "      10 ../testdata/cmd_test/file1.txt\n",
		},
		{
			name:      "wc with -w with matches (extra spaces here and there)",
			path:      []string{"../testdata/cmd_test/file3.txt"},
			countWord: true,
			expResult: "      10 ../testdata/cmd_test/file3.txt\n",
		},
		{
			name:      "wc -c with no matches",
			path:      []string{"../testdata/cmd_test/file2.txt"},
			countChar: true,
			expResult: "       0 ../testdata/cmd_test/file2.txt\n",
		},
		{
			name:      "wc -c with matches",
			path:      []string{"../testdata/cmd_test/file1.txt"},
			countChar: true,
			expResult: "      56 ../testdata/cmd_test/file1.txt\n",
		},
		{
			name:      "wc -c with matches (extra spaces here and there)",
			path:      []string{"../testdata/cmd_test/file3.txt"},
			countChar: true,
			expResult: "      65 ../testdata/cmd_test/file3.txt\n",
		},
		{
			name:      "wc -lwc with matches",
			path:      []string{"../testdata/cmd_test/file1.txt"},
			countLine: true,
			countWord: true,
			countChar: true,
			expResult: "       5       10       56 ../testdata/cmd_test/file1.txt\n",
		},
		{
			name:      "wc no options with matches",
			path:      []string{"../testdata/cmd_test/file1.txt"},
			countLine: false,
			countWord: false,
			countChar: false,
			expResult: "       5       10       56 ../testdata/cmd_test/file1.txt\n",
		},
		{
			name:      "wc multiple files and no options",
			path:      []string{"../testdata/cmd_test/file3.txt", "../testdata/cmd_test/file4.txt"},
			countLine: false,
			countWord: false,
			countChar: false,
			expResult: "       5       10       65 ../testdata/cmd_test/file3.txt\n       7       10       76 ../testdata/cmd_test/file4.txt\n      12       20      141 total\n",
		},
		{
			name:      "wc with stdin",
			path:      []string{},
			stdin:     []byte("xyz abc"),
			countLine: false,
			countWord: false,
			countChar: false,
			expResult: "       0        2        7 \n",
		},
		{
			name:       "wc files with exclude options",
			path:       []string{"../testdata/cmd_test/file3.txt", "../testdata/cmd_test/file4.txt", "../testdata/cmd_test/file5.xyz"},
			countLine:  false,
			countWord:  false,
			countChar:  false,
			excludeExt: []string{"xyz"},
			expResult:  "       5       10       65 ../testdata/cmd_test/file3.txt\n       7       10       76 ../testdata/cmd_test/file4.txt\n      12       20      141 total\n",
		},
		{
			name:       "wc files with include options",
			path:       []string{"../testdata/cmd_test/file3.txt", "../testdata/cmd_test/file4.txt", "../testdata/cmd_test/file5.xyz"},
			countLine:  false,
			countWord:  false,
			countChar:  false,
			includeExt: []string{"txt"},
			expResult:  "       5       10       65 ../testdata/cmd_test/file3.txt\n       7       10       76 ../testdata/cmd_test/file4.txt\n      12       20      141 total\n",
		},
	}

	// creates a file for permission error case, and deletes it in cleanup
	cleanup, err := setTestForPermissonCase(t, "../testdata/cmd_test/file5.txt", "test for permisson case")
	if err != nil {
		t.Fatalf("Unexpected error while setting up test: %v", err)
	}
	defer cleanup()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var got, err bytes.Buffer
			in := bytes.NewReader(tc.stdin)

			input := &WcInput{
				files:      tc.path,
				lineCount:  tc.countLine,
				wordCount:  tc.countWord,
				charCount:  tc.countChar,
				stdin:      in,
				stdout:     &got,
				stderr:     &err,
				excludeExt: tc.excludeExt,
				includeExt: tc.includeExt,
			}

			_ = run(os.DirFS("/"), input)
			want := tc.expResult

			if tc.expError != nil {
				if err.String() == "" {
					t.Fatalf("Expected error but didn't get one\n")
				}

				if !strings.Contains(err.String(), tc.expError.Error()) {
					t.Fatalf("Expected error %q but got %q\n", tc.expError.Error(), err.String())
				}

				return
			}

			if got.String() != want {
				t.Fatalf("Expected %q but got %q\n", want, got.String())
			}
		})
	}
}

func setTestForPermissonCase(t *testing.T, filePath, content string) (func() error, error) {
	t.Helper()

	// creating file
	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}

	// writing to file
	if _, err := file.WriteString(content); err != nil {
		return nil, err
	}

	// setting permisson to 0000
	if err := os.Chmod(filePath, 0000); err != nil {
		return nil, err
	}

	// deletes the file after test
	cleanup := func() error {
		file.Close()
		if err := os.Remove(filePath); err != nil {
			return err
		}
		return nil
	}

	return cleanup, nil
}

func TestPrintResult(t *testing.T) {
	var buffOut, buffErr bytes.Buffer

	tt := []struct {
		name            string
		result          []wc.WcResult
		input           WcInput
		expOut          bool
		expTextInStdOut string
		expTextInStdErr string
	}{
		{
			name: "normal case/case with no errors",
			result: []wc.WcResult{
				{
					Path:      "qwerty",
					LineCount: 8,
					WordCount: 4,
					CharCount: 6,
				},
				{
					Path:      "asdfg",
					LineCount: 4,
					WordCount: 5,
					CharCount: 7,
				},
				{
					Path:      "total",
					LineCount: 12,
					WordCount: 9,
					CharCount: 13,
				},
			},
			input: WcInput{
				lineCount: true,
				wordCount: true,
				charCount: true,
				stdout:    &buffOut,
				stderr:    &buffErr,
			},
			expOut:          false,
			expTextInStdOut: "       8        4        6 qwerty\n       4        5        7 asdfg\n      12        9       13 total\n",
			expTextInStdErr: "",
		},
		{
			name: "case with one error",
			result: []wc.WcResult{
				{
					Path:      "qwerty",
					LineCount: 8,
					WordCount: 4,
					CharCount: 6,
				},
				{
					Path:      "asdfg",
					LineCount: 4,
					WordCount: 5,
					CharCount: 7,
				},
				{
					Err: errors.New("sample error here"),
				},
				{
					Path:      "total",
					LineCount: 12,
					WordCount: 9,
					CharCount: 13,
				},
			},
			input: WcInput{
				lineCount: true,
				wordCount: true,
				charCount: true,
				stdout:    &buffOut,
				stderr:    &buffErr,
			},
			expOut:          true,
			expTextInStdOut: "       8        4        6 qwerty\n       4        5        7 asdfg\n      12        9       13 total\n",
			expTextInStdErr: "sample error here\n",
		},
		{
			name: "case with only error",
			result: []wc.WcResult{
				{
					Err: errors.New("sample error here"),
				},
			},
			input: WcInput{
				lineCount: true,
				wordCount: true,
				charCount: true,
				stdout:    &buffOut,
				stderr:    &buffErr,
			},
			expOut:          true,
			expTextInStdOut: "",
			expTextInStdErr: "sample error here\n",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := printResult(tc.result, &tc.input)

			// check the return value of the function
			if got != tc.expOut {
				t.Errorf("Expected %v but got %v", tc.expOut, got)
			}

			// check the relevant writer for the expected output
			if buffOut.String() != tc.expTextInStdOut {
				t.Errorf("Expected %q but got %q", tc.expTextInStdOut, buffOut.String())
			}
			if buffErr.String() != tc.expTextInStdErr {
				t.Errorf("Expected %q but got %q", tc.expTextInStdErr, buffErr.String())
			}

			// clearing both buffers to be used in next test
			buffOut.Reset()
			buffErr.Reset()
		})
	}
}

// TODO: try to get a case where an error is thrown
func TestGetFullPath(t *testing.T) {
	tt := []struct {
		name   string
		fSys   fs.FS
		path   string
		expOut string
		expErr error
	}{
		{
			name:   "normal case with valid path",
			path:   "testdata/file1.txt",
			expOut: "testdata/file1.txt",
			expErr: nil,
		},
		{
			name:   "normal case with valid path with a lot of back and forth",
			path:   "pkg/../testdata/testdir/../file4.txt",
			expOut: "testdata/file4.txt",
			expErr: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			wd, err := os.Getwd()
			if err != nil {
				t.Fatalf("Unexpected error while getting current working directory: %v", err)
			}

			gotPath, gotErr := getFullPath(os.DirFS(wd), tc.path)
			if tc.expErr != nil {
				if gotErr == nil {
					t.Fatalf("Expected error %q but got nil", tc.expErr.Error())
				}

				if !errors.Is(err, tc.expErr) {
					t.Fatalf("Expected error %q but got %q", tc.expErr.Error(), err.Error())
				}

				return
			}

			if gotErr != nil {
				t.Errorf("Unexpected error occurred: %v", gotErr)
			}

			if gotPath != tc.expOut {
				t.Errorf("Expected %q but got %q", tc.expOut, gotPath)
			}
		})
	}
}
