package scholar

import (
	"fmt"
	"io/ioutil"
	"sort"

	yaml "gopkg.in/yaml.v2"
)

var EntryTypes map[string]*EntryType

// LoadTypes loads the configuration file for entry types and associated
// fields.
func LoadTypes(file string) error {
	d, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(d, &EntryTypes)
	if err != nil {
		return err
	}

	for name, entry := range EntryTypes {
		entry.Type = name
	}

	return nil
}

// TypesInfo shows the information of each entry type. level indicates how much
// information to show (0 = only labels, 1 = labels and required fields, 2 =
// labels, required fields, and optional fields.)
func TypesInfo(level int) {
	var eNames []string
	for name := range EntryTypes {
		eNames = append(eNames, name)
	}
	sort.Strings(eNames)
	for _, name := range eNames {
		EntryTypes[name].info(level)
		fmt.Println()
	}
}

// NewEntry returns an empty copy of an entry according to the types of entries
// loaded.
func NewEntry(label string) (*Entry, error) {
	if entryType, ok := EntryTypes[label]; ok {
		// Return the entry type only if it exists
		return entryType.get(), nil
	}
	// Otherwise, send a type not found error
	TypeNotFoundError = &tnfError{label}
	return nil, TypeNotFoundError
}
