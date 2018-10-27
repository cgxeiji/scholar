package scholar

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type entryType struct {
	Type        string
	Description string            `yaml:"desc"`
	Required    map[string]string `yaml:"req"`
	Optional    map[string]string `yaml:"opt"`
}

func (e *entryType) get() *Entry {
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

func (e *entryType) Info(level int) {
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

type Entry struct {
	Type     string            `yaml:"type"`
	Key      string            `yaml:"key"`
	Required map[string]string `yaml:"req"`
	Optional map[string]string `yaml:"opt"`
}

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

func (e *Entry) Year() string {
	return e.Required["date"][:4]
}

func (e *Entry) FirstAuthorLast() string {
	return strings.Split(e.Required["author"], ",")[0]
}

func (e *Entry) GetKey() string {
	if e.Key == "" {
		e.Key = fmt.Sprintf("%s%s", strings.ToLower(e.FirstAuthorLast()), e.Year())
	}
	return e.Key
}

func (e *Entry) Bib() string {
	bib := fmt.Sprintf("@%s{%s,\n", e.Type, e.GetKey())
	for field, value := range e.Required {
		if value != "" {
			bib = fmt.Sprintf("%s  %s = {%s},\n", bib, field, value)
		}
	}
	for field, value := range e.Optional {
		if value != "" {
			bib = fmt.Sprintf("%s  %s = {%s},\n", bib, field, value)
		}
	}
	bib = fmt.Sprintf("%s}", bib)
	return bib
}
