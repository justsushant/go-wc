/*
Copyright © 2023 justsushant

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "wc",
	Short: "command line program that implements Unix wc like functionality",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) { 
		fs := os.DirFS("/")

		lineCount, _ := cmd.Flags().GetBool("line")
		wordCount, _ := cmd.Flags().GetBool("word")
		charCount, _ := cmd.Flags().GetBool("char")

		err := run(fs, args, lineCount, wordCount, charCount, cmd.InOrStdin() ,cmd.OutOrStdout(), cmd.ErrOrStderr())

		if !err {
			os.Exit(1)
		}
		os.Exit(0)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.word-count.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// rootCmd.Args = cobra.ExactArgs(1)

	rootCmd.PersistentFlags().BoolP("line", "l", false, "show line count")
	rootCmd.PersistentFlags().BoolP("word", "w", false, "show word count")
	rootCmd.PersistentFlags().BoolP("char", "c", false, "show char count")
}