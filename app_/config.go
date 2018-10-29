package main

import (
	"bytes"
	"fmt"

	"github.com/cgxeiji/scholar"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var configDefault = []byte(`
# General Settings
GENERAL:
# Set the default library.
# Scholar will retrieve and save entries from this library.
    default: scholar

# Path locations for the libraries.
# You can add as many libraries as you want.
# You can name the library however you want.
LIBRARIES:
    scholar: ~/ScholarLibrary
`)

func config() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.config/scholar")

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
		fmt.Println("Using default values")
		viper.ReadConfig(bytes.NewBuffer(configDefault))
	}

	dl := viper.Sub("LIBRARIES").GetString(
		viper.GetString("GENERAL.default"))

	dlex, err := homedir.Expand(dl)
	if err != nil {
		panic(err)
	}

	viper.Set("deflib", dlex)

	fmt.Println(viper.GetString("deflib"))

	et := viper.New()
	et.SetConfigName("types")
	et.SetConfigType("yaml")
	et.AddConfigPath(".")
	et.AddConfigPath("$HOME/.config/scholar")

	err = et.ReadInConfig()
	if err != nil {
		fmt.Println(err)
		fmt.Println("Please, set an entry types file at any of those locations.")
		panic("no types.yaml found")
	}

	fmt.Println(et.ConfigFileUsed())
	err = scholar.LoadTypes(et.ConfigFileUsed())
	if err != nil {
		panic(err)
	}

}
