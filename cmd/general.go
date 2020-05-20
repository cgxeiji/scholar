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
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/cgxeiji/crossref"
	"github.com/cgxeiji/scholar/scholar"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

func edit(entry *scholar.Entry) {
	key := entry.GetKey()
	saveTo := filepath.Join(libraryPath(), key)

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

	checkDirKey(libraryPath(), key, entry)
}

func update(entry *scholar.Entry) {
	key := entry.GetKey()
	saveTo := filepath.Join(libraryPath(), key)

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
	case "darwin":
		cmd = "open"
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

func libraryPath() string {
	if currentLibrary != "" {
		if !viper.Sub("LIBRARIES").IsSet(currentLibrary) {
			fmt.Println("No library called", currentLibrary, "was found!")
			fmt.Println("Available libraries:")
			for k, v := range viper.GetStringMapString("LIBRARIES") {
				fmt.Println(" ", k)
				fmt.Println("   ", v)
			}
			panic("not found: library")
		}

		return viper.Sub("LIBRARIES").GetString(currentLibrary)
	}
	return viper.Sub("LIBRARIES").GetString(viper.GetString("GENERAL.default"))
}

func isInteractive() bool {
	return viper.GetBool("GENERAL.interactive") != viper.GetBool("interactive")
}

func entryList() []*scholar.Entry {
	path := libraryPath()
	dirs, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Println(err)
		fmt.Println(`
Add an entry to create this directory or run:

	scholar config

to set the correct path of this library.`,
		)
		panic("not found: library path")
	}

	var wg sync.WaitGroup

	entries := []*scholar.Entry{}
	queue := make(chan *scholar.Entry)
	done := make(chan bool)

	go func() {
		defer close(done)
		for e := range queue {
			entries = append(entries, e)
		}
	}()

	wg.Add(len(dirs))
	for _, dir := range dirs {
		dir := dir
		go func() {
			defer wg.Done()
			if dir.IsDir() {
				filename := filepath.Join(path, dir.Name(), "entry.yaml")
				d, err := ioutil.ReadFile(filename)
				if err != nil {
					fmt.Println("Could not find data for:", dir.Name())
					return
				}
				var e scholar.Entry
				if err := yaml.Unmarshal(d, &e); err != nil {
					panic(err)
				}

				checkDirKey(path, dir.Name(), &e)

				info, err := os.Stat(filename)
				if err != nil {
					fmt.Println("Could not find metadata for:", dir.Name())
					return
				}
				e.Info = info

				queue <- &e
			}
		}()
	}
	wg.Wait()
	close(queue)
	<-done

	return entries
}

// checkDirKey makes sure the directory name is the same as the entry's key.
func checkDirKey(path, dir string, e *scholar.Entry) {
	if dir == e.GetKey() {
		return
	}
	if err := os.Rename(filepath.Join(path, dir), filepath.Join(path, ".tmp.scholar")); err != nil {
		panic(err)
	}

	e.Key = getUniqueKey(e.GetKey())

	if err := os.Rename(filepath.Join(path, ".tmp.scholar"), filepath.Join(path, e.GetKey())); err != nil {
		panic(err)
	}
	update(e)
	fmt.Println("Renamed:")
	fmt.Println(" ", filepath.Join(path, dir), ">",
		filepath.Join(path, e.GetKey()))
}

func queryEntry(search string) *scholar.Entry {
	var entry *scholar.Entry

	if viper.GetBool("GENERAL.interactive") != viper.GetBool("interactive") {
		entry = guiQuery(entryList(), search)
	} else {
		found := guiSearch(search, entryList(), searcher)
		switch len(found) {
		case 0:
			panic("no entries found")
		case 1:
			entry = found[0]
		default:
			panic(fmt.Errorf("too many entries (%d) matched\nplease, refine your query", len(found)))
		}
	}

	return entry
}

func parseCrossref(work *crossref.Work) *scholar.Entry {
	var e *scholar.Entry
	var err error

	switch work.Type {
	case "journal-article":
		if e, err = scholar.NewEntry("article"); err != nil {
			panic(err)
		}
		e.Required["journaltitle"] = work.BookTitle
		e.Optional["issn"] = work.ISSN
	case "proceedings-article":
		if e, err = scholar.NewEntry("inproceedings"); err != nil {
			panic(err)
		}
		e.Required["booktitle"] = work.BookTitle
		e.Optional["isbn"] = work.ISBN
		e.Optional["publisher"] = work.Publisher
	default:
		if e, err = scholar.NewEntry("article"); err != nil {
			panic(err)
		}
		e.Required["journaltitle"] = work.BookTitle
	}

	for _, a := range work.Authors {
		e.Required["author"] = fmt.Sprintf("%s%s, %s and ", e.Required["author"], a.Last, a.First)
	}
	e.Required["author"] = strings.TrimSuffix(e.Required["author"], " and ")

	e.Required["date"] = formatDate(work.Date)
	e.Required["title"] = work.Title

	for _, a := range work.Editors {
		e.Optional["editor"] = fmt.Sprintf("%s%s, %s and ", e.Optional["editor"], a.Last, a.First)
	}
	e.Optional["editor"] = strings.TrimSuffix(e.Optional["editor"], " and ")

	e.Optional["volume"] = work.Volume
	e.Optional["pages"] = work.Pages
	e.Optional["number"] = work.Issue
	e.Optional["doi"] = work.DOI
	e.Optional["abstract"] = work.Abstract

	return e
}

func formatDate(date string) string {
	parts := strings.Split(date, "-")
	for i := range parts {
		fixed := ""
		p, err := strconv.Atoi(parts[i])
		if err != nil {
			panic(err)
		}
		if p < 10 {
			fixed += "0"
		}
		fixed += strconv.Itoa(p)
		parts[i] = fixed
	}

	return strings.Join(parts, "-")
}
