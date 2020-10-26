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
	"strings"

	"github.com/cgxeiji/scholar/scholar"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export [SEARCH]",
	Short: "Export entries",
	Long: `Scholar: a CLI Reference Manager

Print all entries to stdout using biblatex format.

To save all entries to a file run:

	scholar export > references.bib

To specify which entries to export run:

	scholar export SEARCH TERM

--------------------------------------------------------------------------------
TODO: add more export formats
--------------------------------------------------------------------------------
`,
	Run: func(cmd *cobra.Command, args []string) {
		export(args)
	},
}

var exportFormat string

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "biblatex", "Specify the export format (avail: biblatex, bibtex, ris)")
}

func export(args []string) {
	if len(args) != 0 {
		if found := guiSearch(strings.Join(args, " "), entryList(), searcher); len(found) != 0 {
			for _, e := range found {
				fmt.Println(e.Export(exportFormat))
				if exportFormat != "ris" {
					fmt.Println()
				}
			}
		}
		return
	}

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

			fmt.Println(e.Export(exportFormat))
			if exportFormat != "ris" {
				fmt.Println()
			}
		}
	}
}
