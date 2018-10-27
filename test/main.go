package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cgxeiji/crossref"
	"github.com/cgxeiji/scholar"
	"gopkg.in/yaml.v2"
)

var folder string = "ScholarTest"

func AddDOI(es *scholar.Entries, doi string) {
	client := crossref.NewClient("Scholar", "mail@example.com")

	work, err := client.Works(doi)
	if err != nil {
		panic(err)
	}

	entry := es.Parse("crossref", work.Type)

	for _, a := range work.Authors {
		entry.Required["author"] = fmt.Sprintf("%s%s, %s and ", entry.Required["author"], a.Last, a.First)
	}
	entry.Required["author"] = strings.TrimSuffix(entry.Required["author"], " and ")
	entry.Required["date"] = work.Date
	entry.Required["title"] = work.Titles[0]

	if entry.Type == "inproceedings" {
		entry.Required["booktitle"] = work.BookTitles[0]
	} else {
		entry.Required["journaltitle"] = work.BookTitles[0]
	}

	entry.Optional["volume"] = work.Volume
	entry.Optional["number"] = work.Issue
	entry.Optional["doi"] = work.DOI

	key := entry.GetKey()
	saveTo := filepath.Join(folder, key)

	// TODO: check for unique key and directory names

	err = os.MkdirAll(saveTo, os.ModePerm)
	if err != nil {
		panic(err)
	}

	d, err := yaml.Marshal(entry)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(d))
	ioutil.WriteFile(filepath.Join(saveTo, "entry.yaml"), d, 0644)

	var en scholar.Entry
	yaml.Unmarshal(d, &en)
	fmt.Println(en.Bib())
	en.Check()
}

func Add(es *scholar.Entries, entryType string) {

	entry := es.Map[entryType].Get()

	reader := bufio.NewReader(os.Stdin)
	for field, _ := range entry.Required {
		fmt.Printf("%v: ", field)
		text, _ := reader.ReadString('\n')
		text = strings.Trim(text, " \n")
		entry.Required[field] = text
	}

	// entry.Required["author"] = "Last, First"
	// entry.Required["date"] = "2010-12"
	// entry.Required["title"] = "The Title"
	// entry.Required["journaltitle"] = "A Journal of Something"

	// entry.Optional["editor"] = "Editing Company"

	key := entry.GetKey()
	saveTo := filepath.Join(folder, key)

	// TODO: check for unique key and directory names

	err := os.MkdirAll(saveTo, os.ModePerm)
	if err != nil {
		panic(err)
	}

	d, err := yaml.Marshal(entry)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(d))
	ioutil.WriteFile(filepath.Join(saveTo, "entry.yaml"), d, 0644)

	var en scholar.Entry
	yaml.Unmarshal(d, &en)
	fmt.Println(en.Bib())
	en.Check()
}

func Export() {
	dirs, err := ioutil.ReadDir(folder)
	if err != nil {
		panic(err)
	}

	for _, dir := range dirs {
		if dir.IsDir() {
			d, err := ioutil.ReadFile(filepath.Join(folder, dir.Name(), "entry.yaml"))
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

func main() {
	fPrintEntryTypes := flag.Bool("types", false, "Show available entry types")
	fPrintEntryLevel := flag.Int("level", 0, "Set the level of information to be shown")
	fAdd := flag.String("add", "", "Add a new entry")
	fExport := flag.Bool("export", false, "Export entries to biblatex")

	flag.Parse()

	entries := &scholar.Entries{}

	err := entries.Load("types.yaml")
	if err != nil {
		panic(err)
	}

	if *fPrintEntryTypes {
		entries.Show(*fPrintEntryLevel)
	}

	if *fAdd != "" {
		// Add(entries, *fAdd)
		AddDOI(entries, "http://dx.doi.org/10.1016/0004-3702(89)90008-8")
		AddDOI(entries, "http://dx.doi.org/10.1117/12.969296")
	}

	if *fExport {
		Export()
	}

}
