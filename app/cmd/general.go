package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/cgxeiji/scholar"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

func edit(entry *scholar.Entry) {
	key := entry.GetKey()
	saveTo := filepath.Join(viper.GetString("deflib"), key)

	file := filepath.Join(saveTo, "entry.yaml")

	err := editor(file)
	if err != nil {
		panic(err)
	}

	d, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	yaml.Unmarshal(d, &entry)
}

func update(entry *scholar.Entry) {
	key := entry.GetKey()
	saveTo := filepath.Join(viper.GetString("deflib"), key)

	file := filepath.Join(saveTo, "entry.yaml")

	d, err := yaml.Marshal(entry)
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile(file, d, 0644)
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

func open(file string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	default:
		cmd = "xdg-open"
	}
	args = append(args, file)

	return exec.Command(cmd, args...).Start()
}

func clean(filename string) string {
	rx, err := regexp.Compile("[^[:alnum:][:space:]]+")
	if err != nil {
		return filename
	}

	filename = rx.ReplaceAllString(filename, " ")
	filename = strings.Replace(filename, " ", "_", -1)

	return strings.ToLower(filename)
}

func entryQuery(search string) *scholar.Entry {
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
		Selected: `Entry selected: {{ index .Required "title" | cyan | bold }}`,
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
		Label:             "-------------------------- Entries --------------------------",
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
