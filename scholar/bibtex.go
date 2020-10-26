package scholar

import (
	"fmt"
	"sort"
	"strings"
)

type exBibtex struct {
	dict map[string]string
}

func (ex *exBibtex) parse(v string) string {
	if s, ok := ex.dict[v]; ok {
		return s
	}
	return v
}

func (ex *exBibtex) export(e *Entry) string {
	bib := new(strings.Builder)

	// @type{key,
	fmt.Fprintf(bib, "@%s{%s", ex.parse(e.Type), e.GetKey())

	urltmp := ""

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
			switch field {
			case "date":
				date := strings.Split(value, "-")
				fmt.Fprintf(bib, ",\n  %s = {%s}", "year", date[0])
				if len(date) > 1 {
					fmt.Fprintf(bib, ",\n  %s = {%s}", "month", monthText[date[1]])
				}
			case "url":
				urltmp = fmt.Sprintf("\\textsc{url:} \\url{%s}", value) + urltmp
			case "urldate":
				urltmp += fmt.Sprintf(" (accessed %s)", value)
			default:
				field = ex.parse(field)
				fmt.Fprintf(bib, ",\n  %s = {%s}", field, value)
			}
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
			switch field {
			case "url":
				urltmp = fmt.Sprintf("\\textsc{url:} \\url{%s}", value) + urltmp
			case "urldate":
				urltmp += fmt.Sprintf(" (accessed %s)", value)
			default:
				field = ex.parse(field)
				fmt.Fprintf(bib, ",\n  %s = {%s}", field, value)
			}
		}
	}
	if urltmp != "" {
		fmt.Fprintf(bib, ",\n  %s = {%s}", "howpublished", urltmp)
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

var monthText = map[string]string{
	"01": "jan",
	"02": "feb",
	"03": "mar",
	"04": "apr",
	"05": "may",
	"06": "jun",
	"07": "jul",
	"08": "aug",
	"09": "sep",
	"10": "oct",
	"11": "nov",
	"12": "dec",
}

var bibtex = &exBibtex{
	dict: map[string]string{
		"report":       "techreport",
		"online":       "misc",
		"patent":       "misc",
		"langid":       "language",
		"location":     "address",
		"journaltitle": "journal",
		"institution":  "school",
		"date":         "year",
		"url":          "howpublished",
		"urldate":      "howpublished",
	},
}
