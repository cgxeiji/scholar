package scholar

import (
	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/cgxeiji/crossref"
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
func NewEntry(label string) *Entry {
	if entryType, ok := EntryTypes[label]; ok {
		// Return the entry type only if it exists
		return entryType.get()
	}
	// Otherwise, default to misc type
	return EntryTypes["misc"].get()
}

// Parse parses the information for a work to an entry.
func Parse(work *crossref.Work) *Entry {
	return parseCrossref(work)
}

func parseCrossref(work *crossref.Work) *Entry {
	var e *Entry

	switch work.Type {
	case "journal-article":
		e = NewEntry("article")
		e.Required["journaltitle"] = work.BookTitle
		e.Optional["issn"] = work.ISSN
	case "proceedings-article":
		e = NewEntry("inproceedings")
		e.Required["booktitle"] = work.BookTitle
		e.Optional["isbn"] = work.ISBN
		e.Optional["publisher"] = work.Publisher
	default:
		e = NewEntry("article")
		e.Required["journaltitle"] = work.BookTitle
	}

	for _, a := range work.Authors {
		e.Required["author"] = fmt.Sprintf("%s%s, %s and ", e.Required["author"], a.Last, a.First)
	}
	e.Required["author"] = strings.TrimSuffix(e.Required["author"], " and ")

	e.Required["date"] = work.Date
	e.Required["title"] = work.Title

	for _, a := range work.Editors {
		e.Optional["editor"] = fmt.Sprintf("%s%s, %s and ", e.Optional["editor"], a.Last, a.First)
	}
	e.Optional["editor"] = strings.TrimSuffix(e.Optional["editor"], " and ")

	e.Optional["volume"] = work.Volume
	e.Optional["number"] = work.Issue
	e.Optional["doi"] = work.DOI

	return e
}
