// Copyright © 2018 Eiji Onchi <eiji@onchi.me>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove an entry",
	Long: `Scholar: a CLI Reference Manager

Remove an entry from the library.  If interactive mode is disabled, the entry
will be removed without confirmation.
`,
	Run: func(cmd *cobra.Command, args []string) {
		if entry := queryEntry(args); entry != nil {
			path := filepath.Join(libraryPath(), entry.GetKey())
			if viper.GetBool("GENERAL.interactive") != viper.GetBool("interactive") {
				if askYesNo(fmt.Sprintf("Do you want to remove %s?", path)) {
					if err := os.RemoveAll(path); err != nil {
						panic(err)
					}
					fmt.Println("Removed", path)
				}
			} else {
				path := filepath.Join(libraryPath(), entry.GetKey())
				if err := os.RemoveAll(path); err != nil {
					panic(err)
				}
				fmt.Println("Removed", path)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
