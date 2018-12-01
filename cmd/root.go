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
	"path/filepath"
	"runtime"

	"github.com/cgxeiji/scholar/scholar"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var confFile, typesFile, currentLibrary string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "scholar",
	Short: "A CLI Reference Manager",
	Long: `Scholar: a CLI Reference Manager

Scholar is a CLI reference manager that keeps track of
your documents metadata using YAML files with biblatex format.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(os.Args[1:]) == 0 {
			cmd.Help()
		}
		return
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("error:", r)
			os.Exit(1)
		}
	}()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&currentLibrary, "library", "l", "", "specify the library")
	rootCmd.PersistentFlags().BoolP("interactive", "i", false, "toggle interactive mode (enabled by default)")
	viper.BindPFlag("interactive", rootCmd.PersistentFlags().Lookup("interactive"))
}

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

	// Check if there are configuration files
	viper.AutomaticEnv()

	// Set default values
	viper.SetDefault("GENERAL.interactive", true)
	viper.SetDefault("GENERAL.editor", "vi")

	// Load the configuration file. If not found, auto-generate one.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println(err)
		fmt.Println("Setting up a new configuration file")

		path, _ := homedir.Dir()
		path = filepath.Join(path, ".config", "scholar")
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			panic(err)
		}

		path = filepath.Join(path, "config.yaml")
		switch runtime.GOOS {
		case "windows":
			if err := ioutil.WriteFile(path, configTemplateWin, 0644); err != nil {
				panic(err)
			}
		default:
			if err := ioutil.WriteFile(path, configTemplate, 0644); err != nil {
				panic(err)
			}
		}

		if err := viper.ReadInConfig(); err != nil {
			panic(err)
		}
	}

	if confFile == "which" {
		fmt.Println("Configuration file used:", viper.ConfigFileUsed())
	}

	for k, v := range viper.GetStringMapString("LIBRARIES") {
		vex, err := homedir.Expand(v)
		if err != nil {
			panic(err)
		}
		viper.Set("LIBRARIES."+k, vex)
	}

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

	// Load the configuration file. If not found, auto-generate one.
	if err := et.ReadInConfig(); err != nil {
		fmt.Println(err)
		fmt.Println("Setting up a new types file")

		path, _ := homedir.Dir()
		path = filepath.Join(path, ".config", "scholar")
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			panic(err)
		}

		path = filepath.Join(path, "types.yaml")
		if err := ioutil.WriteFile(path, typesTemplate, 0644); err != nil {
			panic(err)
		}

		if err := et.ReadInConfig(); err != nil {
			panic(err)
		}
	}

	if typesFile == "which" {
		fmt.Println("Types file used:", et.ConfigFileUsed())
	}

	err := scholar.LoadTypes(et.ConfigFileUsed())
	if err != nil {
		panic(err)
	}
}
