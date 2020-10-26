package scholar

import (
	"fmt"
	"sort"
	"strings"
)

type exBiblatex struct{}

func (ex *exBiblatex) export(e *Entry) string {
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

var biblatex = &exBiblatex{}
