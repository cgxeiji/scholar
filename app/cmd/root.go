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
	"bytes"
	"fmt"
	"os"

	"github.com/cgxeiji/scholar"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var confFile, typesFile, curentLibrary string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "scholar",
	Short: "A CLI Reference Manager",
	Long: `Scholar: a CLI Reference Manager

Scholar is a CLI reference manager that keeps track of
your documents metadata using YAML files with biblatex format.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		return
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&confFile, "config", "", "config file (default $HOME/.config/scholar/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&typesFile, "types", "", "entry types file (default $HOME/.config/scholar/types.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

var configDefault = []byte(`
# General Settings
GENERAL:
    # Set the default library.
    # Scholar will retrieve and save entries from this library.
    default: scholar
    # Set the default text editor
    editor: vi
    # Set the email for polite use of CrossRef
    mailto: mail@example.com

# Path locations for the libraries.
# You can add as many libraries as you want.
# You can name the library however you want.
LIBRARIES:
    scholar: ~/ScholarLibrary
`)

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if confFile != "" && confFile != "which" {
		// Use config file from the flag.
		viper.SetConfigFile(confFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.config/scholar")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println(err)
		fmt.Println("Using default values")
		viper.ReadConfig(bytes.NewBuffer(configDefault))
	}

	if confFile == "which" {
		fmt.Println("Configuration file used:", viper.ConfigFileUsed())
	}

	dl := viper.Sub("LIBRARIES").GetString(
		viper.GetString("GENERAL.default"))

	dlex, err := homedir.Expand(dl)
	if err != nil {
		panic(err)
	}

	viper.Set("deflib", dlex)

	et := viper.New()
	if typesFile != "" && typesFile != "which" {
		// Use config file from the flag.
		et.SetConfigFile(typesFile)
	} else {
		et.SetConfigName("types")
		et.SetConfigType("yaml")
		et.AddConfigPath(".")
		et.AddConfigPath("$HOME/.config/scholar")
	}

	err = et.ReadInConfig()
	if err != nil {
		fmt.Println(err)
		fmt.Println("Please, set an entry types file at any of those locations.")
		panic("no types.yaml found")
	}

	if typesFile == "which" {
		fmt.Println("Types file used:", et.ConfigFileUsed())
	}

	err = scholar.LoadTypes(et.ConfigFileUsed())
	if err != nil {
		panic(err)
	}
}
