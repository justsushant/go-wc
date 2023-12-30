package cmd

import (
	"fmt"
	"testing"
	"testing/fstest"
)

func TestRun(t *testing.T) {
	file1Name := "file1.txt"
	file1Data := []byte("this\nis\na\nmulti line\ntext")

	testFS := fstest.MapFS{
		file1Name: {Data: file1Data, Mode: 0755},
	}

	testCases := []struct{
		name string
		lineCount bool
		wordCount bool
		charCount bool
		output string
	}{
		{"only line count", true, false, false, fmt.Sprintf("       5 %s", file1Name)},
		{"only word count", false, true, false, fmt.Sprintf("       6 %s", file1Name)},
		{"only char count", false, false, true, fmt.Sprintf("      25 %s", file1Name)},
		{"all counts", true, true, true, fmt.Sprintf("       5        6       25 %s", file1Name)},
		{"combination of line & char count", true, false, true, fmt.Sprintf("       5       25 %s", file1Name)},
		
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := Run(testFS, file1Name, tc.lineCount, tc.wordCount, tc.charCount)
			want := tc.output

			if got != want {
				t.Errorf("Expected %q but got %q", want, got)
			}
		})
	}
}