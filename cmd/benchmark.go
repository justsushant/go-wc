package cmd

import (
	"os"
	"testing"
)

func BenchmarkRun(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// wd, _ := os.Getwd()
		fSys := os.DirFS("/")

		input := &WcInput{
			files:     []string{"../testdata/loadtest/file.txt"},
			lineCount: true,
			stdout:    os.Stdout,
			stderr:    os.Stdout,
		}

		_ = run(fSys, input)

	}
}
