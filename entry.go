package scholar

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// EntryType defines how each entry will be formatted. Each entry has
// a TYPE of entry, a short DESCRIPTION, REQUIRED fields, and
// OPTIONAL fields according to BibLaTex documentation.
type EntryType struct {
	Type        string
	Description string            `yaml:"desc"`
	Required    map[string]string `yaml:"req"`
	Optional    map[string]string `yaml:"opt"`
}

func (e *EntryType) get() *Entry {
	var c Entry
	c.Type = e.Type
	c.Required = make(map[string]string)
	for k := range e.Required {
		c.Required[k] = ""
	}

	c.Optional = make(map[string]string)
	for k := range e.Optional {
		c.Optional[k] = ""
	}

	return &c
}

func (e *EntryType) info(level int) {
	fmt.Println(e.Type, ":", e.Description)

	if level > 0 {
		var fields []string
		for f := range e.Required {
			fields = append(fields, f)
		}
		sort.Strings(fields)
		for _, field := range fields {
			fmt.Println("  ", field, "->", e.Required[field])
		}

		if level > 1 {
			fields = nil
			for f := range e.Optional {
				fields = append(fields, f)
			}
			sort.Strings(fields)
			for _, field := range fields {
				fmt.Printf("     (%v) -> %v\n", field, e.Optional[field])
			}
		}
	}
}

// Entry is the basic object of scholar.
type Entry struct {
	Type     string            `yaml:"type"`
	Key      string            `yaml:"key"`
	Required map[string]string `yaml:"req"`
	Optional map[string]string `yaml:"opt"`
	File     string            `yaml:"file"`
}

// Attach attaches a file path to the entry.
func (e *Entry) Attach(file string) {
	e.File = file
}

// Check checks if the fields are formatted correctly.
// [Currently not useful]
func (e *Entry) Check() error {
	date := e.Required["date"]
	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		_, err = time.Parse("2006-01", date)
		if err != nil {
			_, err = time.Parse("2006", date)
			if err != nil {
				return fmt.Errorf("invalid date format (date %s). Please use YYYY[-MM[-DD]]", date)
			}
		}
	}

	return nil
}

// Year returns the year of the entry.
func (e *Entry) Year() string {
	return fmt.Sprintf("%.4s", e.Required["date"])
}

// FirstAuthorLast return the lastname of the first author of the entry.
func (e *Entry) FirstAuthorLast() string {
	return strings.Split(e.Required["author"], ",")[0]
}

// GetKey return the key of the entry. If there is no key, a new key is
// generated with lastnameYEAR format.
// For example: einstein1922
func (e *Entry) GetKey() string {
	if e.Key == "" {
		e.Key = fmt.Sprintf("%s%s", strings.ToLower(e.FirstAuthorLast()), e.Year())
	}
	return e.Key
}

// Convert changes the type of an entry, parsing all the fields from one
// type to the other.
// If the entry does not exits, it defaults to misc entry.
// TODO: create a ErrFieldNotFound error
func Convert(e *Entry, entryType string) *Entry {
	to := NewEntry(entryType)
	to.Key = e.Key
	to.Attach(e.File)

	// Check for the new required fields in the required and
	// optional fields of the old entry
	for field := range to.Required {
		if value, ok := e.Required[field]; ok {
			to.Required[field] = value
			delete(e.Required, field)
		} else if value, ok := e.Optional[field]; ok {
			to.Required[field] = value
			delete(e.Optional, field)
		}
	}

	// Dump of remaining required fields of the old entry to
	// optional fields of the new entry
	for field, value := range e.Required {
		if value != "" {
			to.Optional[field] = value
		}
	}
	for field, value := range e.Optional {
		if value != "" {
			to.Optional[field] = value
		}
	}

	return to
}

// Bib returns a string with all the information of the entry
// in BibLaTex format.
func (e *Entry) Bib() string {
	bib := fmt.Sprintf("@%s{%s,\n", e.Type, e.GetKey())
	var fields []string
	for f := range e.Required {
		fields = append(fields, f)
	}
	sort.Strings(fields)
	for _, field := range fields {
		if value := e.Required[field]; value != "" {
			bib = fmt.Sprintf("%s  %s = {%s},\n", bib, field, value)
		}
	}

	fields = nil
	for f := range e.Optional {
		fields = append(fields, f)
	}
	sort.Strings(fields)
	for _, field := range fields {
		if value := e.Optional[field]; value != "" && field != "abstract" {
			bib = fmt.Sprintf("%s  %s = {%s},\n", bib, field, value)
		}
	}
	if value, ok := e.Optional["abstract"]; ok {
		bib = fmt.Sprintf("%s  %s = {%s},\n", bib, "abstract", value)
	}

	bib = fmt.Sprintf("%s\n}", bib[:len(bib)-2])
	return bib
}
