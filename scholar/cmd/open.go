// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
		open(entryQuery("").File)
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
