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
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/cgxeiji/scholar"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

// openCmd represents the open command
var openCmd = &cobra.Command{
	Use:   "open [KEY]",
	Short: "Opens an entry",
	Long: `Scholar: a CLI Reference Manager

Open an entry's attached file with the default system's software.

To search for an entry run:

	scholar open

to specify which entry to open run:

	scholar open KEY

--------------------------------------------------------------------------------
TODO: if there are multiple files attached, a selection menu appears.
TODO: if there is no file attached, the entry's metadata file is opened.
--------------------------------------------------------------------------------
`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Make a entry search menu
		open(queryFrom(entryList(), strings.Join(args, " ")).File)
	},
}

func init() {
	rootCmd.AddCommand(openCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// openCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// openCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func entryFromKey(key string) *scholar.Entry {
	dirs, err := ioutil.ReadDir(libraryPath())
	if err != nil {
		panic(err)
	}

	for _, dir := range dirs {
		if dir.IsDir() && dir.Name() == strings.TrimSpace(key) {
			d, err := ioutil.ReadFile(filepath.Join(libraryPath(), dir.Name(), "entry.yaml"))
			if err != nil {
				panic(err)
			}

			var e scholar.Entry
			err = yaml.Unmarshal(d, &e)
			if err != nil {
				panic(err)
			}

			return &e
		}
	}

	return &scholar.Entry{}
}
