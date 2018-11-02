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
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/cgxeiji/scholar"
	"github.com/jroimartin/gocui"
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
			os.Exit(1)
		}

		return viper.Sub("LIBRARIES").GetString(currentLibrary)
	}
	return viper.Sub("LIBRARIES").GetString(viper.GetString("GENERAL.default"))
}

func queryFrom(list []*scholar.Entry, search string) *scholar.Entry {
	gui(list, search)
	if selectedEntry == nil {
		os.Exit(0)
	}
	return selectedEntry
}

func gui(entries []*scholar.Entry, search string) {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		panic(err)
	}
	defer g.Close()

	g.Highlight = true
	g.SelFgColor = gocui.ColorGreen | gocui.AttrBold
	g.FgColor = gocui.ColorWhite

	g.SetManagerFunc(
		func(g *gocui.Gui) error {
			maxX, maxY := g.Size()

			if v, err := g.SetView("main", 0, 3, maxX/5*3, maxY-1); err != nil {
				if err != gocui.ErrUnknownView {
					return err
				}
				v.Title = "ENTRIES"
				v.Editable = true
				v.Highlight = true
				v.SelBgColor = gocui.ColorGreen
				v.SelFgColor = gocui.ColorBlack

				for _, e := range entries {
					showList = append(showList, e)
					fmt.Fprint(v, formatEntry(e, maxX/5*3))
				}

				if _, err := g.SetCurrentView("main"); err != nil {
					return err
				}

				v.Editor = gocui.EditorFunc(func(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
					switch {
					case key == gocui.KeyArrowDown:
						v.MoveCursor(0, 1, false)
					case key == gocui.KeyArrowUp:
						v.MoveCursor(0, -1, false)
					case ch == 'j':
						v.MoveCursor(0, 1, false)
					case ch == 'k':
						v.MoveCursor(0, -1, false)
					}

					g.Update(func(g *gocui.Gui) error {
						_, oy := v.Origin()
						_, cy := v.Cursor()

						vd, err := g.View("detail")
						if err != nil {
							panic(err)
						}

						vd.Clear()
						if len(showList) > 0 && cy+oy < len(showList) {
							formatEntryInfo(vd, showList[cy+oy])
						}

						return nil
					})
				})
			}

			if v, err := g.SetView("detail", maxX/5*3+1, 3, maxX-1, maxY-1); err != nil {
				if err != gocui.ErrUnknownView {
					return err
				}
				v.Title = "DETAILS"
				v.Wrap = true
				formatEntryInfo(v, showList[0])
			}

			if v, err := g.SetView("search", 0, 0, maxX-1, 2); err != nil {
				if err != gocui.ErrUnknownView {
					return err
				}
				v.Editable = true
				v.Title = "SEARCH BAR"
				fmt.Fprint(v, search)

				searcher := func(input string, entry *scholar.Entry) bool {
					title := strings.Replace(strings.ToLower(entry.Required["title"]), " ", "", -1)
					aus := strings.Replace(strings.ToLower(entry.Required["author"]), " ", "", -1)
					k := strings.Replace(strings.ToLower(entry.Key), " ", "", -1)
					s := fmt.Sprintf("%s%s%s", title, aus, k)
					input = strings.TrimSpace(input)
					input = strings.Replace(strings.ToLower(input), " ", "", -1)

					return strings.Contains(s, input)
				}

				g.Update(func(g *gocui.Gui) error {
					vm, err := g.View("main")
					if err != nil {
						panic(err)
					}

					vm.SetCursor(0, 0)
					vm.SetOrigin(0, 0)
					guiSearch(v, vm, entries, searcher)
					v.SetCursor(len(search), 0)

					if len(showList) == 1 {
						selectedEntry = showList[0]
						return gocui.ErrQuit
					}

					vd, err := g.View("detail")
					if err != nil {
						panic(err)
					}
					vd.Clear()
					if len(showList) > 0 {
						formatEntryInfo(vd, showList[0])
					}
					return nil
				})

				v.Editor = gocui.EditorFunc(func(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
					switch {
					case ch != 0 && mod == 0:
						v.EditWrite(ch)
					case key == gocui.KeySpace:
						v.EditWrite(' ')
					case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
						v.EditDelete(true)
					case key == gocui.KeyDelete:
						v.EditDelete(false)
					case key == gocui.KeyInsert:
						v.Overwrite = !v.Overwrite
					case key == gocui.KeyArrowLeft:
						v.MoveCursor(-1, 0, false)
					case key == gocui.KeyArrowRight:
						v.MoveCursor(1, 0, false)
					}

					g.Update(func(g *gocui.Gui) error {
						vm, err := g.View("main")
						if err != nil {
							panic(err)
						}

						vm.SetCursor(0, 0)
						vm.SetOrigin(0, 0)
						guiSearch(v, vm, entries, searcher)

						vd, err := g.View("detail")
						if err != nil {
							panic(err)
						}
						vd.Clear()
						if len(showList) > 0 {
							formatEntryInfo(vd, showList[0])
						}
						return nil
					})
				})
			}

			helpString := " /: search, space: biblatex, enter: open, q: quit, ^c: exit "

			if v, err := g.SetView("help", 1, maxY-2, len(helpString)+2, maxY); err != nil {
				if err != gocui.ErrUnknownView {
					return err
				}
				v.Frame = false
				fmt.Fprint(v, helpString)
			}
			return nil
		})

	if err := g.SetKeybinding("main", 'q', gocui.ModNone, quit); err != nil {
		panic(err)
	}

	if err := g.SetKeybinding("main", '/', gocui.ModNone, toggleSearch); err != nil {
		panic(err)
	}

	if err := g.SetKeybinding("search", gocui.KeyEnter, gocui.ModNone, toggleSearch); err != nil {
		panic(err)
	}

	g.InputEsc = true
	if err := g.SetKeybinding("search", gocui.KeyEsc, gocui.ModNone, toggleSearch); err != nil {
		panic(err)
	}

	if err := g.SetKeybinding("main", gocui.KeyEnter, gocui.ModNone, guiSelect); err != nil {
		panic(err)
	}

	if err := g.SetKeybinding("main", gocui.KeySpace, gocui.ModNone, guiShowInfo); err != nil {
		panic(err)
	}

	if err := g.SetKeybinding("info", gocui.KeySpace, gocui.ModNone, guiHideInfo); err != nil {
		panic(err)
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		panic(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		panic(err)
	}
}

func toggleSearch(g *gocui.Gui, v *gocui.View) error {
	if v == nil || v.Name() == "search" {
		_, err := g.SetCurrentView("main")
		g.Cursor = false
		return err
	}
	_, err := g.SetCurrentView("search")
	g.Cursor = true
	return err
}

func guiSelect(g *gocui.Gui, v *gocui.View) error {
	_, cy := v.Cursor()
	selectedEntry = showList[cy]

	return gocui.ErrQuit
}

func guiShowInfo(g *gocui.Gui, v *gocui.View) error {
	_, cy := v.Cursor()
	entry := showList[cy]

	maxX, maxY := g.Size()
	if v, err := g.SetView("info", maxX/10, maxY/10, maxX*9/10, maxY*9/10); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = true
		v.Editable = true
		g.Cursor = true

		v.Editor = gocui.EditorFunc(func(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
			switch {
			case key == gocui.KeyArrowDown:
				v.MoveCursor(0, 1, false)
			case key == gocui.KeyArrowUp:
				v.MoveCursor(0, -1, false)
			case key == gocui.KeyArrowLeft:
				v.MoveCursor(-1, 0, false)
			case key == gocui.KeyArrowRight:
				v.MoveCursor(1, 0, false)
			case ch == 'j':
				v.MoveCursor(0, 1, false)
			case ch == 'k':
				v.MoveCursor(0, -1, false)
			case ch == 'h':
				v.MoveCursor(-1, 0, false)
			case ch == 'l':
				v.MoveCursor(1, 0, false)
			}
		})

		fmt.Fprint(v, entry.Bib())
		if _, err := g.SetCurrentView("info"); err != nil {
			return err
		}
	}

	return nil
}

func guiHideInfo(g *gocui.Gui, v *gocui.View) error {
	if err := g.DeleteView("info"); err != nil {
		return err
	}
	g.Cursor = false
	if _, err := g.SetCurrentView("main"); err != nil {
		return err
	}
	return nil
}

var showList []*scholar.Entry
var selectedEntry *scholar.Entry

func guiSearch(vsearch *gocui.View, vmain *gocui.View, entries []*scholar.Entry, searcher func(string, *scholar.Entry) bool) {
	vmain.Clear()

	input := vsearch.Buffer()
	showList = []*scholar.Entry{}

	for _, e := range entries {
		if searcher(input, e) {
			showList = append(showList, e)
		}
	}

	maxX, _ := vmain.Size()

	for _, e := range showList {
		fmt.Fprint(vmain, formatEntry(e, maxX))
	}

}

func formatEntry(entry *scholar.Entry, width int) string {
	return fmt.Sprintf("\033[32;1m%-*.*s  \033[33;1m(%-4.4s)  \033[31;1m%-*.*s\033[0m\n",
		width/3*2-4, width/3*2-4, entry.Required["title"],
		entry.Required["date"],
		width/3, width/3, entry.Required["author"])
}

func formatEntryInfo(w io.Writer, e *scholar.Entry) {
	fmt.Fprintf(w, "\033[32;7m[%s]\033[0m\n",
		strings.ToTitle(e.Type))
	fmt.Fprintf(w, "Title:\n  \033[32;1m%s\033[0m\n",
		e.Required["title"])
	aus := strings.Split(e.Required["author"], " and ")
	fmt.Fprintf(w, "Author(s):\n")
	for _, au := range aus {
		fmt.Fprintf(w, "  \033[31;1m%s\033[0m\n",
			au)
	}

	fmt.Fprintf(w, "Date:\n  \033[33;1m%s\033[0m\n",
		e.Required["date"])

	var fields []string
	for f := range e.Required {
		if f != "title" && f != "author" && f != "date" {
			fields = append(fields, f)
		}
	}
	if value := e.File; value != "" {
		fmt.Fprintf(w, "%s:\n  \033[31;4m%s\033[0m\n", "File", value)
	}
	sort.Strings(fields)
	for _, field := range fields {
		if value := e.Required[field]; value != "" {
			fmt.Fprintf(w, "%s:\n  \033[33;1m%s\033[0m\n", strings.Title(field), value)
		}
	}

	fields = nil
	for f := range e.Optional {
		fields = append(fields, f)
	}
	sort.Strings(fields)
	for _, field := range fields {
		if value := e.Optional[field]; value != "" && field != "abstract" {
			fmt.Fprintf(w, "%s:\n  \033[33;1m%s\033[0m\n", strings.Title(field), value)
		}
	}
	if value, ok := e.Optional["abstract"]; ok {
		fmt.Fprintf(w, "%s:\n  \033[33;1m%s\033[0m\n", strings.Title("abstract"), value)
	}
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func entryList() []*scholar.Entry {
	path := libraryPath()
	dirs, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}

	var entries []*scholar.Entry

	for _, dir := range dirs {
		if dir.IsDir() {
			d, err := ioutil.ReadFile(filepath.Join(path, dir.Name(), "entry.yaml"))
			if err != nil {
				panic(err)
			}

			var e scholar.Entry
			err = yaml.Unmarshal(d, &e)
			if err != nil {
				panic(err)
			}

			checkDirKey(path, dir.Name(), &e)

			entries = append(entries, &e)
		}
	}

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
