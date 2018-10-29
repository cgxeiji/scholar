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
	"strings"

	"github.com/cgxeiji/crossref"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add [FILENAME/QUERY]",
	Short: "Adds a new entry",
	Long: `Add a new entry to scholar.

You can TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("add called")
		fmt.Printf("with %v arguments\n", args)
		query(strings.Join(args, " "))
	},
}

var addDoi, addAttach string

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().StringVarP(&addDoi, "doi", "d", "", "Specify the DOI to retrieve metadata")
	addCmd.Flags().StringVar(&curentLibrary, "to", "", "specify library to use")
	addCmd.Flags().StringVarP(&addAttach, "attach", "a", "", "attach a file to the entry")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func query(search string) {
	fmt.Printf("Searching for: %s\n", search)

	client := crossref.NewClient("Scholar", "mail@example.com")

	ws, err := client.Query(search)
	if err != nil {
		panic(err)
	}

	for _, v := range ws {
		au := "Unknown"
		if len(v.Authors) > 0 {
			au = fmt.Sprintf("%s, %s", v.Authors[0].Last, v.Authors[0].First)
		}
		fmt.Printf("  > %-20.20s | %-10.10s | %4.4s | %20.20s\n", v.Title, au, v.Date, v.DOI)
	}
}
