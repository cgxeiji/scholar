package scholar

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
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

func (e *EntryType) info(level int) string {
	b := new(bytes.Buffer)

	b.WriteString(fmt.Sprintf("%s: %s\n", e.Type, e.Description))

	if level > 0 {
		fields := make([]string, len(e.Required))
		i := 0
		for field := range e.Required {
			fields[i] = field
			i++
		}
		sort.Strings(fields)
		for _, field := range fields {
			b.WriteString(fmt.Sprintf("  %s -> %s\n", field, e.Required[field]))
		}

		if level > 1 {
			fields := make([]string, len(e.Optional))
			i := 0
			for field := range e.Optional {
				fields[i] = field
				i++
			}
			sort.Strings(fields)
			for _, field := range fields {
				b.WriteString(fmt.Sprintf("     (%v) -> %v\n", field, e.Optional[field]))
			}
		}
	}

	return b.String()
}

// String implements the Stringer interface.
func (e *EntryType) String() string {
	return fmt.Sprintf("%q (req: %d, opt: %d)", e.Type, len(e.Required), len(e.Optional))
}

// Entry is the basic object of scholar.
type Entry struct {
	Type     string            `yaml:"type"`
	Key      string            `yaml:"key"`
	Required map[string]string `yaml:"req"`
	Optional map[string]string `yaml:"opt"`
	File     string            `yaml:"file"`
	Info     os.FileInfo       `yaml:"-"`
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

// Convert changes the type of an entry, parsing all the fields from one type
// to the other.
//
// If the entry type does not exits, it returns with an ErrTypeNotFound error.
// If a required field from the output entry type cannot be found on the
// original entry, the Required field is replaced by an empty string and
// Convert keeps parsing the entry. An ErrFieldNotFound error is raised as a
// result, but it is safe to ignore ErrFieldNotFound.
//
// Any field that was Required in the original entry, but it is not on the
// output entry type, is converted to an Optional field. This ensures back and
// forth conversion of the same entry without losing information.
func Convert(e *Entry, entryType string) (*Entry, error) {
	to, err := NewEntry(entryType)
	if err != nil {
		return to, getError("Convert", ErrTypeNotFound, err)
	}
	to.Key = e.Key
	to.Attach(e.File)

	seen := make(map[string]bool)
	var convertError error

	// Check for the new required fields in the required and
	// optional fields of the old entry
	for field := range to.Required {
		if value, ok := e.Required[field]; ok {
			to.Required[field] = value
		} else if value, ok := e.Optional[field]; ok {
			to.Required[field] = value
		} else {
			convertError = getError("Convert", ErrFieldNotFound, convertError).
				info(fmt.Sprintf("required field %s[%s] was not found in entry of type %q, replacing with an empty string",
					to.Type, field, e.Type))
			to.Required[field] = ""
		}
		seen[field] = true
	}

	// Dump of remaining required fields of the old entry to
	// optional fields of the new entry
	for field, value := range e.Required {
		if seen[field] {
			continue
		}
		if value != "" {
			to.Optional[field] = value
		}
		seen[field] = true
	}
	for field, value := range e.Optional {
		if seen[field] {
			continue
		}
		if value != "" {
			to.Optional[field] = value
		}
		seen[field] = true
	}

	return to, convertError
}

// Bib returns a string with all the information of the entry
// in BibLaTex format.
func (e *Entry) Bib() string {
	bib := new(strings.Builder)

	// @type{key,
	fmt.Fprintf(bib, "@%s{%s", e.Type, e.GetKey())

	// field = {value},
	fields := make([]string, len(e.Required))
	i := 0
	for field := range e.Required {
		fields[i] = field
		i++
	}
	sort.Strings(fields)
	for _, field := range fields {
		if value := e.Required[field]; value != "" {
			fmt.Fprintf(bib, ",\n  %s = {%s}", field, value)
		}
	}

	fields = make([]string, len(e.Optional))
	i = 0
	for field := range e.Optional {
		fields[i] = field
		i++
	}
	sort.Strings(fields)
	for _, field := range fields {
		if value := e.Optional[field]; value != "" && field != "abstract" {
			fmt.Fprintf(bib, ",\n  %s = {%s}", field, value)
		}
	}
	if value, ok := e.Optional["abstract"]; ok {
		fmt.Fprintf(bib, ",\n  %s = {%s}", "abstract", value)
	}
	if file := e.File; file != "" {
		fmt.Fprintf(bib, ",\n  %s = {%s}", "file", file)
	}

	bib.WriteString("\n}")

	return bib.String()
}

// Export returns a string with all the information of the entry in the given
// format.
func (e *Entry) Export(format string) string {
	ex := getExporter(format)

	return ex.export(e)
}

const bibTemplate = `@[[ .Type ]]{[[ .GetKey ]]
[[- range $field, $value := .Required -]]
  [[ if $value -]]
    [[ ",\n  " ]][[ $field ]] = {[[ $value ]]}
  [[- end ]]
[[- end ]]

[[- range $field, $value := .Optional -]]
  [[ if $value -]]
    [[ ",\n  " ]][[ $field ]] = {[[ $value ]]}
  [[- end ]]
[[- end ]]
}`

var tmpl = template.Must(
	template.New("biblatex").
		Delims("[[", "]]").
		Parse(bibTemplate),
)

// bibT returns a string with all the information of the entry
// in BibLaTex format using a text template.
//
// Currently testing performance.
func (e *Entry) bibT() string {
	bib := new(strings.Builder)

	err := tmpl.Execute(bib, e)
	if err != nil {
		panic(err)
	}

	return bib.String()
}
