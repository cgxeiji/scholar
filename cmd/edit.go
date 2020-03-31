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

	"github.com/cgxeiji/scholar/scholar"
	"github.com/spf13/cobra"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit an entry",
	Long: `Scholar: a CLI Reference Manager

Edit an entry's metadata using the default's text editor.
`,
	Run: func(cmd *cobra.Command, args []string) {
		if entry := queryEntry(args); entry != nil {
			if attachFlag != "" {
				attach(entry, attachFlag)
				return
			}
			if editType != "" {
				var err error
				entry, err = scholar.Convert(entry, editType)
				if err != nil {
					panic(err)
				}
				update(entry)
			}
			edit(entry)
		} else {
			panic("entry not found")
		}
	},
}

var editType string

func init() {
	rootCmd.AddCommand(editCmd)

	editCmd.Flags().StringVarP(&attachFlag, "attach", "a", "", "attach a file to the entry")
	editCmd.Flags().StringVarP(&editType, "type", "t", "", "change the type of the entry")
}
