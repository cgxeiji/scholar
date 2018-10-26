package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/cgxeiji/scholar"
	"gopkg.in/yaml.v2"
)

var folder string = "ScholarTest"

func AddDOI(es *scholar.Entries, doi string) {
	fmt.Println(fmt.Sprintf("https://api.crossref.org/v1/works/%s", doi))
	resp, err := http.Get(fmt.Sprintf("https://api.crossref.org/v1/works/%s", doi))
	//resp, err := http.Get("https://google.com")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var data map[string]interface{}
	json.Unmarshal(raw, &data)

	// fmt.Println(data)

	content, _ := data["message"].(map[string]interface{})
	fmt.Println(content["type"])
	fmt.Println("Title:", content["title"])
	fmt.Println("In:", content["container-title"])

	fmt.Println("Author:")
	authors := reflect.ValueOf(content["author"])
	for i := 0; i < authors.Len(); i++ {
		author := authors.Index(i).Interface().(map[string]interface{})
		fmt.Println("  ", author["given"], author["family"])
	}
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
	}

	if *fExport {
		Export()
	}

}
