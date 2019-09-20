package scholar

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"sort"

	yaml "gopkg.in/yaml.v2"
)

// EntryTypes holds the list of all entry types loaded.
var EntryTypes map[string]*EntryType

// LoadTypes loads the configuration file for entry types and associated
// fields.
func LoadTypes(file string) error {
	d, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	return loadTypes(d)
}

func loadTypes(b []byte) error {
	err := yaml.Unmarshal(b, &EntryTypes)
	if err != nil {
		return err
	}

	for name, entry := range EntryTypes {
		entry.Type = name
	}

	return nil
}

// TypesInfo shows the information of each entry type. The level indicates how
// much information to show (0 = only labels, 1 = labels and required fields, 2
// = labels, required fields, and optional fields.)
func TypesInfo(level int) {
	fmt.Print(typesInfo(0))
}

// FTypesInfo writes the information of each entry type to the io.Writer w. The
// level indicates how much information to show (0 = only labels, 1 = labels
// and required fields, 2 = labels, required fields, and optional fields.)
func FTypesInfo(w io.Writer, level int) {
	fmt.Fprint(w, typesInfo(level))
}

func typesInfo(level int) string {
	b := new(bytes.Buffer)
	eNames := make([]string, len(EntryTypes))
	i := 0
	for name := range EntryTypes {
		eNames[i] = name
		i++
	}
	sort.Strings(eNames)
	for _, name := range eNames {
		s := EntryTypes[name].info(level)
		b.WriteString(s)
		b.WriteString("\n")
	}

	return b.String()
}

// NewEntry returns an empty copy of an entry according to the types of entries
// loaded.
func NewEntry(label string) (*Entry, error) {
	if entryType, ok := EntryTypes[label]; ok {
		// Return the entry type only if it exists
		return entryType.get(), nil
	}
	// Otherwise, send a type not found error
	keys := make([]string, len(EntryTypes))
	i := 0
	for k := range EntryTypes {
		keys[i] = k
		i++
	}

	return nil,
		getError("NewEntry", ErrTypeNotFound, nil).
			info(fmt.Sprintf("%q is not a valid entry type, available types: %q", label, keys))
}
