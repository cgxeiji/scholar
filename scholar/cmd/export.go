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
	"io/ioutil"
	"path/filepath"

	"github.com/cgxeiji/scholar"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Exports entries",
	Long: `Scholar: a CLI Reference Manager

Print all entries to stdout using biblatex format.

To save to a file run:

	scholar export > references.bib

--------------------------------------------------------------------------------
TODO: add different export formats
TODO: add flag to specify which library to export
--------------------------------------------------------------------------------
`,
	Run: func(cmd *cobra.Command, args []string) {
		export()
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// exportCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// exportCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func export() {
	path := libraryPath()
	if currentLibrary != "" {
		path = viper.Sub("LIBRARIES").GetString(currentLibrary)
	}
	dirs, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}

	for _, dir := range dirs {
		if dir.IsDir() {
			d, err := ioutil.ReadFile(filepath.Join(path, dir.Name(), "entry.yaml"))
			if err != nil {
				panic(err)
			}

			var e scholar.Entry
			err = yaml.Unmarshal(d, &e)
			if err != nil {
				panic(err)
			}

			fmt.Println(e.Bib())
			fmt.Println()
		}
	}
}
