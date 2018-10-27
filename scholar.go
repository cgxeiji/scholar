package scholar

import (
	"fmt"
	"io/ioutil"
	"sort"

	yaml "gopkg.in/yaml.v2"
)

var entryTypes map[string]*entryType

// LoadTypes loads the configuration file for entry types and associated
// fields.
func LoadTypes(file string) error {
	d, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(d, &entryTypes)
	if err != nil {
		return err
	}

	for name, entry := range entryTypes {
		entry.Type = name
	}

	return nil
}

// TypesInfo shows the information of each entry type. level indicates how much
// information to show (0 = only labels, 1 = labels and required fields, 2 =
// labels, required fields, and optional fields.)
func TypesInfo(level int) {
	var eNames []string
	for name := range entryTypes {
		eNames = append(eNames, name)
	}
	sort.Strings(eNames)
	for _, name := range eNames {
		entryTypes[name].Info(level)
		fmt.Println()
	}
}

// NewEntry returns an empty copy of an entry according to the types of entries
// loaded.
func NewEntry(service, label string) *Entry {
	switch service {
	case "crossref":
		switch label {
		case "journal-article":
			return entryTypes["article"].get()
		case "proceedings-article":
			return entryTypes["inproceedings"].get()
		default:
			return entryTypes["article"].get()
		}
	case "none":
		return entryTypes[label].get()
	}

	return &Entry{}
}
