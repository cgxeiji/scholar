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
	"strings"

	"github.com/cgxeiji/scholar/scholar"
	"github.com/jroimartin/gocui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// uiCmd represents the ui command
var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		guiScholar(entryList(), "")
	},
}

func init() {
	rootCmd.AddCommand(uiCmd)
}

func guiScholar(entries []*scholar.Entry, search string) *scholar.Entry {
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
			fmt.Fprint(v, search)

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

			v.SetCursor(len(search), 0)
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
					if found := guiSearch(v.Buffer(), entries, searcher); len(found) > 0 {
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
			openEntry(showList[oy+cy])
			return nil
		}); err != nil {
		panic(err)
	}

	if err := g.SetKeybinding("main", 'e', gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			_, oy := v.Origin()
			_, cy := v.Cursor()
			edit(showList[oy+cy])
			// For now, until it is possible to redraw the GUI
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
