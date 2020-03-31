// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
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
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   "fetch [SEARCH]",
	Short: "Prints the file path of the entry",
	Long: `Scholar: a CLI Reference Manager

Fetch the file path attached to an entry and print the result to stdout.
If no file is attached, it exists with 1.
`,
	Run: func(cmd *cobra.Command, args []string) {
		if entry := queryEntry(args); entry != nil {
			if entry.File != "" {
				fmt.Println(filepath.Join(libraryPath(), entry.GetKey(), entry.File))
			} else if url, ok := entry.Optional["url"]; ok && url != "" {
				fmt.Println(url)
			} else if url, ok := entry.Required["url"]; ok && url != "" {
				fmt.Println(url)
			} else if doi, ok := entry.Optional["doi"]; ok && doi != "" {
				fmt.Println(fmt.Sprintf("https://dx.doi.org/%s", doi))
			} else {
				panic("no file, doi, or url associated with entry")
			}
		} else {
			panic("entry not found")
		}
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)
}
