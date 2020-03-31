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
	"sort"
	"strings"
	"sync"

	"github.com/cgxeiji/scholar/scholar"
	"github.com/jroimartin/gocui"
	"github.com/spf13/viper"
)

var helpString = " /: search, s: sort, space: cite, enter: select, q: quit, ^c: exit "
var sortby = []string{"modified", "title", "author", "date"}
var sortid = 0
var showList []*scholar.Entry

func guiQuery(entries []*scholar.Entry, search []string) *scholar.Entry {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		panic(err)
	}
	defer g.Close()

	showInfoCh := make(chan *scholar.Entry)
	defer close(showInfoCh)
	selEntryCh := make(chan *scholar.Entry, 1)
	// this channel is closed before returning the entry
	resetCursorCh := make(chan bool)
	defer close(resetCursorCh)

	g.Highlight = true
	g.SelFgColor = gocui.ColorGreen | gocui.AttrBold
	g.FgColor = gocui.ColorWhite

	g.SetManagerFunc(func(g *gocui.Gui) error {
		maxX, maxY := g.Size()

		if v, err := g.SetView("main", 0, 3, maxX/5*3, maxY-1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}

			// Initialize main view
			l := viper.GetString("GENERAL.default")
			if currentLibrary != "" {
				l = currentLibrary
			}
			v.Title = fmt.Sprintf("ENTRIES:%s", strings.ToTitle(l))
			v.Editable = true
			v.Highlight = true
			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack

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

				_, oy := v.Origin()
				_, cy := v.Cursor()
				if len(showList) > 0 && cy+oy < len(showList) {
					showInfoCh <- showList[cy+oy]
				}
			})

			go func() {
				for range resetCursorCh {
					v.SetCursor(0, 0)
					v.SetOrigin(0, 0)
				}
			}()
			// End initialization

			if _, err := g.SetCurrentView("main"); err != nil {
				return err
			}
		} else {
			v.Clear()
			for i, e := range showList {
				fe := formatEntry(e, maxX*3/5)
				if i == len(showList)-1 {
					fe = strings.TrimSpace(fe)
				}
				fmt.Fprint(v, fe)
			}
		}

		if v, err := g.SetView("detail", maxX/5*3+1, 3, maxX-1, maxY-1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = "DETAILS"
			v.Wrap = true

			go func() {
				for e := range showInfoCh {
					g.Update(func(g *gocui.Gui) error {
						v.Clear()
						if e != nil {
							formatEntryInfo(v, e)
						}
						return nil
					})
				}
			}()
		}

		if v, err := g.SetView("search", 0, 0, maxX-1, 2); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Editable = true
			v.Title = "SEARCH BAR"
			search_string := strings.Join(search, " ")
			fmt.Fprint(v, search_string)

			// Check if the initial search is a unique result
			found := guiSearch(search, entries, searcher)
			switch len(found) {
			case 0:
				return gocui.ErrQuit
			case 1:
				selEntryCh <- found[0]
				return gocui.ErrQuit
			default:
				showInfoCh <- found[0]
			}

			v.SetCursor(len(search_string), 0)
			resetCursorCh <- true

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

				if ch != 0 && mod == 0 || key == gocui.KeyBackspace || key == gocui.KeyBackspace2 || key == gocui.KeyDelete {
					if found := guiSearch(strings.Split(v.Buffer(), " "), entries, searcher); len(found) > 0 {
						showInfoCh <- found[0]
					} else {
						showInfoCh <- nil
					}
					resetCursorCh <- true
				}
			})
		}

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

	if err := g.SetKeybinding("main", 's', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		sortid = (sortid + 1) % len(sortby)
		guiSort(showList, sortby[sortid])
		if len(showList) > 0 {
			_, oy := v.Origin()
			_, cy := v.Cursor()
			showInfoCh <- showList[oy+cy]
		}
		return nil
	}); err != nil {
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

	if err := g.SetKeybinding("main", gocui.KeyEnter, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			_, oy := v.Origin()
			_, cy := v.Cursor()
			selEntryCh <- showList[oy+cy]
			return gocui.ErrQuit
		}); err != nil {
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

	close(selEntryCh)
	return <-selEntryCh
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

func guiShowInfo(g *gocui.Gui, v *gocui.View) error {
	_, oy := v.Origin()
	_, cy := v.Cursor()
	entry := showList[oy+cy]

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

func guiSearch(search []string, entries []*scholar.Entry, searcher func(string, *scholar.Entry) bool) []*scholar.Entry {
	var wg sync.WaitGroup

	found := []*scholar.Entry{}
	queue := make(chan *scholar.Entry)
	done := make(chan bool)

	go func() {
		defer close(done)
		for e := range queue {
			found = append(found, e)
		}
	}()

	wg.Add(len(entries))
	for _, e := range entries {
		e := e
		go func() {
			defer wg.Done()

			// Check intersection of match with each key
			match := true
			for _,key := range search {
				match = match && searcher(key, e)
			}
			if match {
				queue <- e
			}
		}()
	}
	wg.Wait()
	close(queue)
	<-done

	// TODO: find a better way to share indexed list
	guiSort(found, sortby[sortid])
	showList = found

	return found
}

func guiSort(entries []*scholar.Entry, field string) {
	if field == "modified" {
		sort.SliceStable(entries, func(i, j int) bool {
			return entries[i].Info.ModTime().After(entries[j].Info.ModTime())
		})
		return
	}

	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].Required[field] < entries[j].Required[field]
	})
}

func formatEntry(entry *scholar.Entry, width int) string {
	return fmt.Sprintf("\033[32;1m%-*.*s  \033[33;1m(%-4.4s)  \033[31;1m%-*.*s\033[0m\n",
		width/3*2-4, width/3*2-4, entry.Required["title"],
		entry.Required["date"],
		width/3, width/3, entry.Required["author"])
}

func formatEntryInfo(w io.Writer, e *scholar.Entry) {
	fmt.Fprintf(w, "\033[32;7m[%s]\033[0m \033[31;4m%s\033[0m\n",
		strings.ToTitle(e.Type),
		e.GetKey())
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

func searcher(input string, entry *scholar.Entry) bool {
	title := strings.Replace(strings.ToLower(entry.Required["title"]), " ", "", -1)
	aus := strings.Replace(strings.ToLower(entry.Required["author"]), " ", "", -1)
	k := strings.Replace(strings.ToLower(entry.Key), " ", "", -1)
	s := fmt.Sprintf("%s%s%s", title, aus, k)
	input = strings.TrimSpace(input)
	input = strings.Replace(strings.ToLower(input), " ", "", -1)

	return strings.Contains(s, input)
}
