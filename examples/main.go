package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cgxeiji/crossref"
	"github.com/cgxeiji/scholar"
	"github.com/skratchdot/open-golang/open"
	"gopkg.in/yaml.v2"
)

var folder = "~/ScholarTest"

func addDOI(doi string) *scholar.Entry {
	client := crossref.NewClient("Scholar", "mail@example.com")

	work, err := client.Works(doi)
	if err != nil {
		panic(err)
	}

	entry := scholar.NewEntry("crossref", work.Type)

	for _, a := range work.Authors {
		entry.Required["author"] = fmt.Sprintf("%s%s, %s and ", entry.Required["author"], a.Last, a.First)
	}
	entry.Required["author"] = strings.TrimSuffix(entry.Required["author"], " and ")

	for _, a := range work.Editors {
		entry.Optional["editor"] = fmt.Sprintf("%s%s, %s and ", entry.Optional["editor"], a.Last, a.First)
	}
	entry.Optional["editor"] = strings.TrimSuffix(entry.Optional["editor"], " and ")
	entry.Required["date"] = work.Date
	entry.Required["title"] = work.Title

	if entry.Type == "inproceedings" {
		entry.Required["booktitle"] = work.BookTitle
		entry.Optional["isbn"] = work.ISBN
		entry.Optional["publisher"] = work.Publisher
	} else {
		entry.Required["journaltitle"] = work.BookTitle
		entry.Optional["issn"] = work.ISSN
	}

	entry.Optional["volume"] = work.Volume
	entry.Optional["number"] = work.Issue
	entry.Optional["doi"] = work.DOI

	return entry
}

func commit(entry *scholar.Entry) {
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

	file := filepath.Join(saveTo, "entry.yaml")
	ioutil.WriteFile(file, d, 0644)

	open.Run(file)

	d, err = ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	var en scholar.Entry
	yaml.Unmarshal(d, &en)
	fmt.Println(en.Bib())
	en.Check()
}

func attach(entry *scholar.Entry, file string) {
	// horrible placeholder
	key := entry.GetKey()
	saveTo := filepath.Join(folder, key)

	src, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer src.Close()

	path := filepath.Join(saveTo, fmt.Sprintf("%s%s", key, filepath.Ext(file)))

	dst, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer dst.Close()

	b, err := io.Copy(dst, src)
	if err != nil {
		panic(err)
	}
	fmt.Println("Copied", b, "bytes")
	// horrible placeholder
	entry.File = path
}

func add(entryType string) *scholar.Entry {

	entry := scholar.NewEntry("none", entryType)

	reader := bufio.NewReader(os.Stdin)
	for field := range entry.Required {
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

	return entry
}

func export() {
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

func find(key string) *scholar.Entry {
	dirs, err := ioutil.ReadDir(folder)
	if err != nil {
		panic(err)
	}

	for _, dir := range dirs {
		if dir.IsDir() && dir.Name() == strings.TrimSpace(key) {
			d, err := ioutil.ReadFile(filepath.Join(folder, dir.Name(), "entry.yaml"))
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

func modeAdd() {
	flags := flag.NewFlagSet("add", flag.ExitOnError)

	fAttach := flags.String("attach", "", "Copy and attach a file to the entry")
	fDOI := flags.String("doi", "", "Add metadata from DOI")
	fType := flags.String("type", "article", "Specify the type of entry")
	flags.Parse(os.Args[2:])

	e := &scholar.Entry{}

	if *fDOI != "" {
		e = addDOI("http://dx.doi.org/10.1016/0004-3702(89)90008-8")
	} else {
		e = add(*fType)
	}

	if *fAttach != "" {
		attach(e, *fAttach)
	}

	commit(e)
}

func main() {
	config()
	return

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "add":
			modeAdd()
			return
		}
	}

	fPrintEntryTypes := flag.Bool("types", false, "Show available entry types")
	fPrintEntryLevel := flag.Int("level", 0, "Set the level of information to be shown")
	fAdd := flag.String("add", "", "Add a new entry")
	fExport := flag.Bool("export", false, "Export entries to biblatex")
	fAttach := flag.String("attach", "", "Copy and attach a file to the entry")
	fOpen := flag.String("open", "", "Open an entry (key)")

	flag.Parse()

	if *fPrintEntryTypes {
		scholar.TypesInfo(*fPrintEntryLevel)
	}

	if *fAdd != "" {
		// Add(entries, *fAdd)
		e := addDOI("http://dx.doi.org/10.1016/0004-3702(89)90008-8")

		if *fAttach != "" {
			attach(e, *fAttach)
		}

		commit(e)
		commit(addDOI("http://dx.doi.org/10.1117/12.969296"))
	}

	if *fExport {
		export()
	}

	if *fOpen != "" {
		e := find(*fOpen)
		if e.File != "" {
			open.Start(e.File)
		}
	}

}
