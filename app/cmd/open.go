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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cgxeiji/scholar"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

// openCmd represents the open command
var openCmd = &cobra.Command{
	Use:   "open",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Make a entry search menu
		open(findQuery("").File)
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

func findFromKey(key string) *scholar.Entry {
	dirs, err := ioutil.ReadDir(viper.GetString("deflib"))
	if err != nil {
		panic(err)
	}

	for _, dir := range dirs {
		if dir.IsDir() && dir.Name() == strings.TrimSpace(key) {
			d, err := ioutil.ReadFile(filepath.Join(viper.GetString("deflib"), dir.Name(), "entry.yaml"))
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

func findQuery(search string) *scholar.Entry {
	dirs, err := ioutil.ReadDir(viper.GetString("deflib"))
	if err != nil {
		panic(err)
	}

	var entries []*scholar.Entry

	for _, dir := range dirs {
		if dir.IsDir() {
			d, err := ioutil.ReadFile(filepath.Join(viper.GetString("deflib"), dir.Name(), "entry.yaml"))
			if err != nil {
				panic(err)
			}

			var e scholar.Entry
			err = yaml.Unmarshal(d, &e)
			if err != nil {
				panic(err)
			}

			entries = append(entries, &e)
		}
	}

	template := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   `> {{ index .Required "title" | cyan | bold | underline }} ({{ index .Required "date" | yellow | bold | underline }}) {{ index .Required "author" | red | bold | underline }}`,
		Inactive: `  {{ index .Required "title" | cyan }} ({{ index .Required "date" | yellow }}) {{ index .Required "author" | red }}`,
		Selected: `Entry type: {{ index .Required "title" | cyan | bold }}`,
		Details: `
------------------------- Details -------------------------
{{ "Title:" | faint }}	{{ index .Required "title" | cyan | bold}}
{{ "Author(s):" | faint }}	{{ index .Required "author" | red | bold}}
{{ "Date:" | faint }}	{{ index .Required "date" | yellow | bold}}`,
	}

	searcher := func(input string, index int) bool {
		entry := entries[index]
		title := strings.Replace(strings.ToLower(entry.Required["title"]), " ", "", -1)
		aus := strings.Replace(strings.ToLower(entry.Required["author"]), " ", "", -1)
		file := strings.Replace(strings.ToLower(filepath.Base(entry.File)), "_", "", -1)
		s := fmt.Sprintf("%s%s%s", title, aus, file)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(s, input)
	}

	prompt := promptui.Select{
		Label:             "-------------------------- Types --------------------------",
		Items:             entries,
		Templates:         template,
		Size:              5,
		Searcher:          searcher,
		StartInSearchMode: true,
	}

	i, _, err := prompt.Run()

	if err != nil {
		fmt.Println("Aborting")
		os.Exit(1)
	}

	return entries[i]

}
