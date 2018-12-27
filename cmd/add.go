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
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cgxeiji/crossref"
	"github.com/cgxeiji/scholar/scholar"
	"github.com/manifoldco/promptui"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add [FILENAME/QUERY]",
	Short: "Add a new entry",
	Long: `Scholar: a CLI Reference Manager

Add a new entry to a library.

--------------------------------------------------------------------------------
TODO: Add a flag for manual/auto input of metadata
--------------------------------------------------------------------------------
`,
	Run: func(cmd *cobra.Command, args []string) {
		var entry *scholar.Entry
		var err error
		var file string
		doi := doiFlag

		input := strings.Join(args, " ")
		file, err = homedir.Expand(input)
		if err != nil {
			panic(err)
		}
		if _, err := os.Stat(file); os.IsNotExist(err) {
			file = ""
		} else {
			input = ""
		}

		if doi == "" {
			if input == "" {
				if askYesNo("Would you like to search the web for metadata?") {
					doi = query(requestSearch())
				}
			} else {
				doi = query(input)
			}
		}
		if doi == "" {
			entry = manual()
		} else {
			entry = addDOI(doi)
		}

		commit(entry)
		if file != "" {
			info.println("attaching:", file)
			attach(entry, file)
		}
		edit(entry)

		info.println()
		info.println(entry.Bib())
	},
}

var doiFlag, addAttach string

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().StringVarP(&doiFlag, "doi", "d", "", "Specify the DOI to retrieve metadata")
	addCmd.Flags().StringVarP(&addAttach, "attach", "a", "", "attach a file to the entry")
}

func askYesNo(question string) bool {
	prompt := promptui.Prompt{
		Label:     question,
		IsConfirm: true,
	}

	res, _ := prompt.Run()
	if res == "" {
		os.Exit(1)
	}

	return strings.Contains("yesYes", res)
}

func requestSearch() string {
	prompt := promptui.Prompt{
		Label: "Search for",
	}

	res, err := prompt.Run()

	if err != nil {
		panic("aborting")
	}

	return res
}

func getUniqueKey(key string) string {
	path := libraryPath()
	mark := 'a'
	valid := key

	for _, err := os.Stat(filepath.Join(path, valid)); !os.IsNotExist(err); _, err = os.Stat(filepath.Join(path, valid)) {
		valid = fmt.Sprintf("%s%s", key, string(mark))
		mark++
	}

	return valid
}

func commit(entry *scholar.Entry) {
	entry.Key = getUniqueKey(entry.GetKey())
	saveTo := filepath.Join(libraryPath(), entry.Key)

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
	fmt.Println("  ..", file)
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

	e := parseCrossref(w)

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
	entry, err := scholar.NewEntry(entryType)
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(os.Stdin)
	for field := range entry.Required {
		fmt.Printf("%v: ", field)
		text, _ := reader.ReadString('\n')
		text = strings.Trim(text, " \n")
		entry.Required[field] = text
	}

	return entry
}

func attach(entry *scholar.Entry, file string) {
	key := entry.GetKey()
	saveTo := filepath.Join(libraryPath(), key)

	src, err := os.Open(file)
	if err != nil {
		fmt.Println("Attempted to:")
		fmt.Println(" ", err)
		return
	}
	defer src.Close()

	filename := fmt.Sprintf("%s_%.40s%s", key, clean(entry.Required["title"]), filepath.Ext(file))

	path := filepath.Join(saveTo, filename)

	dst, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer dst.Close()

	b, err := io.Copy(dst, src)
	if err != nil {
		panic(err)
	}
	fmt.Println("Copied", b, "bytes to", path)
	// horrible placeholder
	entry.Attach(filename)

	update(entry)
}

func manual() *scholar.Entry {
	info.println("Adding the entry manually...")
	info.println("Select the type of entry:")
	t := selectType()
	info.println("Required fields:")
	return add(t)
}
