package main

import (
	"fmt"
	"io/ioutil"
	"reflect"

	"gopkg.in/yaml.v2"
)

func main() {
	test := make(map[interface{}]interface{})

	data, err := ioutil.ReadFile("types.yaml")
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(data, &test)
	if err != nil {
		panic(err)
	}

	for entry, ev := range test {
		fmt.Println(entry)
		evv := reflect.ValueOf(ev)
		if evv.Kind() == reflect.Map {
			for _, content := range evv.MapKeys() {
				cv := evv.MapIndex(content).Interface()
				fmt.Println("  ", content)

				cvv := reflect.ValueOf(cv)
				if cvv.Kind() == reflect.Map {
					for _, field := range cvv.MapKeys() {
						fv := cvv.MapIndex(field)
						fmt.Println("    ", field, "->", fv)
					}
				} else {
					fmt.Println("    >", cv)
				}
			}
		}
	}

}
