package cmd

import (
	"bytes"
	"io/fs"
	"os"
	"strings"
	"testing"

	wc "github.com/one2n-go-bootcamp/word-count/pkg"
)

func TestRun(t *testing.T) {
	testCases := []struct{
		name string
		path string
		countLine bool
		countWord bool
		countChar bool
		expResult string
		expError error
	}{
		// {
		// 	name: "wc with -l with matches",
		// 	path: "../testdata/cmd_test/file1.txt",
		// 	countLine: true, 
		// 	expResult: "       6 ../testdata/cmd_test/file1.txt\n",
		// },
		{
			name: "wc over non-existent-file",
			path: "testdata/cmd_test/non-existent-file.txt",
			expError: fs.ErrNotExist,
		},
		{
			name: "wc over directory",
			path: "testdata/cmd_test/testdir",
			expError: wc.ErrIsDirectory,
		},
		{
			name: "wc -l with no matches",
			path: "testdata/cmd_test/file2.txt",
			countLine: true, 
			expResult: "       0 testdata/cmd_test/file2.txt\n",
		},
		{
			name: "wc -l with matches",
			path: "testdata/cmd_test/file1.txt",
			countLine: true, 
			expResult: "       5 testdata/cmd_test/file1.txt\n",
		},
		{
			name: "wc -w with no matches",
			path: "testdata/cmd_test/file2.txt",
			countWord: true, 
			expResult: "       0 testdata/cmd_test/file2.txt\n",
		},
		{
			name: "wc -w with matches",
			path: "testdata/cmd_test/file1.txt",
			countWord: true, 
			expResult: "      10 testdata/cmd_test/file1.txt\n",
		},
		{
			name: "wc with -w with matches (extra spaces here and there)",
			path: "testdata/cmd_test/file3.txt",
			countWord: true, 
			expResult: "      10 testdata/cmd_test/file3.txt\n",
		},
		{
			name: "wc -c with no matches",
			path: "testdata/cmd_test/file2.txt",
			countChar: true, 
			expResult: "       0 testdata/cmd_test/file2.txt\n",
		},
		{
			name: "wc -c with matches",
			path: "testdata/cmd_test/file1.txt",
			countChar: true, 
			expResult: "      56 testdata/cmd_test/file1.txt\n",
		},
		{
			name: "wc -c with matches (extra spaces here and there)",
			path: "testdata/cmd_test/file3.txt",
			countChar: true, 
			expResult: "      65 testdata/cmd_test/file3.txt\n",
		},
		{
			name: "wc -lwc with matches",
			path: "testdata/cmd_test/file1.txt",
			countLine: true,
			countWord: true,
			countChar: true, 
			expResult: "       5       10       56 testdata/cmd_test/file1.txt\n",
		},
		{
			name: "wc no options with matches",
			path: "testdata/cmd_test/file1.txt",
			countLine: false,
			countWord: false,
			countChar: false, 
			expResult: "       5       10       56 testdata/cmd_test/file1.txt\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var got, err bytes.Buffer
			
			_ = run(os.DirFS("../"), []string{tc.path}, tc.countLine, tc.countWord, tc.countChar, &got, &err)
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