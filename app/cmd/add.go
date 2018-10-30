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
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/cgxeiji/crossref"
	"github.com/cgxeiji/scholar"
	"github.com/kr/pretty"
	"github.com/manifoldco/promptui"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add [FILENAME/QUERY]",
	Short: "Adds a new entry",
	Long: `Add a new entry to scholar.

You can TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		var entry *scholar.Entry

		search := strings.Join(args, " ")
		file, _ := homedir.Expand(search)
		if _, err := os.Stat(file); os.IsNotExist(err) {
			if doi := query(search); doi != "" {
				entry = addDOI(doi)
			}
		} else {
			fmt.Println("file:", file)
			s := filepath.Base(file)
			s = strings.TrimSuffix(s, filepath.Ext(s))
			doi := addDoi
			if doi == "" {
				doi = query(s)
			}

			if doi == "" {
				fmt.Println()
				if !requestManual() {
					doi = query(requestSearch())
				}
				if doi != "" {
					fmt.Println("Getting metadata from doi")
					entry = addDOI(doi)
				} else {
					fmt.Println()
					fmt.Println("Adding the entry manually...")
					fmt.Println("What kind of entry is it?")
					t := selectType()
					fmt.Println()
					fmt.Println("Please, add the required fields:")
					entry = add(t)
				}

			}

		}

		commit(entry)
		pretty.Println(entry.Bib())

	},
}

var addDoi, addAttach string

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().StringVarP(&addDoi, "doi", "d", "", "Specify the DOI to retrieve metadata")
	addCmd.Flags().StringVar(&curentLibrary, "to", "", "specify library to use")
	addCmd.Flags().StringVarP(&addAttach, "attach", "a", "", "attach a file to the entry")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func requestManual() bool {
	prompt := promptui.Prompt{
		Label:     "I could not find anything, can you give me a better search variable?",
		IsConfirm: true,
	}

	res, _ := prompt.Run()

	return !strings.Contains("yesYes", res)
}

func requestSearch() string {
	prompt := promptui.Prompt{
		Label: "Search for",
	}

	res, err := prompt.Run()

	if err != nil {
		fmt.Println("Aborting")
		os.Exit(1)
	}

	return res
}

func commit(entry *scholar.Entry) {
	fmt.Println("commiting")
	key := entry.GetKey()
	saveTo := filepath.Join(viper.GetString("deflib"), key)

	if _, err := os.Stat(saveTo); !os.IsNotExist(err) {
		//TODO: make a better algorithm for unique keys
		saveTo = fmt.Sprintf("%sa", saveTo)
		entry.Key = fmt.Sprintf("%sa", key)
	}

	err := os.MkdirAll(saveTo, os.ModePerm)
	if err != nil {
		panic(err)
	}

	d, err := yaml.Marshal(entry)
	if err != nil {
		panic(err)
	}

	file := filepath.Join(saveTo, "entry.yaml")
	ioutil.WriteFile(file, d, 0644)

	err = editor(file)
	if err != nil {
		panic(err)
	}

	d, err = ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	yaml.Unmarshal(d, &entry)
}

func editor(file string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	default:
		cmd = viper.GetString("GENERAL.editor")
	}
	args = append(args, file)
	c := exec.Command(cmd, args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c.Run()
}

func query(search string) string {
	fmt.Println("Searching metadata for:", search)

	client := crossref.NewClient("Scholar", viper.GetString("GENERAL.mailto"))

	ws, err := client.Query(search)
	if err != nil {
		panic(err)
	}

	type work struct {
		Title  string
		Short  string
		Author string
		Year   string
		DOI    string
	}

	works := []work{}

	switch len(ws) {
	case 0:
		fmt.Println("Nothing found...")
		return ""
	case 1:
		return ws[0].DOI
	}

	for _, v := range ws {
		works = append(works, work{
			Title:  v.Title,
			Short:  fmt.Sprintf("%20.20s", v.Title),
			Author: fmt.Sprintf("%v", v.Authors),
			Year:   fmt.Sprintf("%4.4s", v.Date),
			DOI:    v.DOI,
		})
	}

	template := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "> {{ .Short | yellow | bold | underline }} ({{ .Year | yellow | bold | underline }}) {{ .Author | yellow | bold | underline }}",
		Inactive: "  {{ .Short | cyan }} ({{ .Year | yellow }}) {{ .Author | red}}",
		Selected: "Parsing entry for: {{ .Title | cyan | bold }}",
		Details: `
------------------------- Details -------------------------
{{ "Title:" | faint }}	{{ .Title | cyan | bold}}
{{ "Author(s):" | faint }}	{{ .Author | red | bold}}
{{ "Year:" | faint }}	{{ .Year | yellow | bold}}
{{ "DOI:" | faint }}	{{ .DOI | bold }}`,
	}

	searcher := func(input string, index int) bool {
		work := works[index]
		title := strings.Replace(strings.ToLower(work.Title), " ", "", -1)
		authors := strings.Replace(strings.ToLower(work.Author), " ", "", -1)
		s := fmt.Sprintf("%s%s", title, authors)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(s, input)
	}

	fmt.Println()

	prompt := promptui.Select{
		Label:             "-------------------------- Found --------------------------",
		Items:             works,
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

	return works[i].DOI
}

func addDOI(doi string) *scholar.Entry {
	client := crossref.NewClient("Scholar", viper.GetString("GENERAL.mailto"))

	w, err := client.Works(doi)
	if err != nil {
		panic(err)
	}

	e := scholar.Parse(w)

	return e
}

func selectType() string {
	entries := []*scholar.EntryType{}

	var eNames []string
	for name := range scholar.EntryTypes {
		eNames = append(eNames, name)
	}
	sort.Strings(eNames)
	for _, name := range eNames {
		entries = append(entries, scholar.EntryTypes[name])
	}

	template := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "> {{ .Type | yellow | bold | underline }} {{ .Description | cyan | bold | underline }}",
		Inactive: "  {{ .Type | yellow }} {{ .Description | cyan }}",
		Selected: "Entry type: {{ .Type | yellow | bold }}",
		Details: `
------------------------- Details -------------------------
{{ .Type | yellow | bold}}
{{ .Description | cyan | bold}}`,
	}

	searcher := func(input string, index int) bool {
		entry := entries[index]
		title := strings.Replace(strings.ToLower(entry.Type), " ", "", -1)
		desc := strings.Replace(strings.ToLower(entry.Description), " ", "", -1)
		s := fmt.Sprintf("%s%s", title, desc)
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

	return entries[i].Type
}

func add(entryType string) *scholar.Entry {
	entry := scholar.NewEntry(entryType)

	reader := bufio.NewReader(os.Stdin)
	for field := range entry.Required {
		fmt.Printf("%v: ", field)
		text, _ := reader.ReadString('\n')
		text = strings.Trim(text, " \n")
		entry.Required[field] = text
	}

	return entry
}
