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
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure Scholar",
	Long: `Scholar: a CLI Reference Manager

Configure and modify the settings of Scholar.

Scholar will search for a configuration file
in the current directory, or at the default
location.

If there is no configuration file available,
a configuration file will be created at:
	
	$HOME/.config/scholar/config.yaml

--------------------------------------------------------------------------------
TODO: add option to create a local configuration
--------------------------------------------------------------------------------
`,
	Run: func(cmd *cobra.Command, args []string) {
		configure()
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func configure() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.config/scholar")

	// Check if there are configuration files
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println(err)
		fmt.Println("Setting up a new configuration file")

		path, _ := homedir.Dir()
		path = filepath.Join(path, ".config", "scholar", "config.yaml")
		if err := ioutil.WriteFile(path, configTemplate, 0644); err != nil {
			panic(nil)
		}

		if err := viper.ReadInConfig(); err != nil {
			panic(nil)
		}
	}

	editor(viper.ConfigFileUsed())
}
