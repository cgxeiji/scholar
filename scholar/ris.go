package scholar

import (
	"fmt"
	"strings"
)

type exRIS struct {
	dict map[string]string
}

func (ex *exRIS) parse(v string) string {
	if s, ok := ex.dict[v]; ok {
		return s
	}
	return ""
}

func (ex *exRIS) export(e *Entry) string {
	ris := new(strings.Builder)

	// TY  - type
	fmt.Fprintf(ris, "TY  - %s", ex.parse(e.Type))

	// field  - value
	for field := range e.Required {
		if value := e.Required[field]; value != "" {
			switch field {
			case "author":
				authors := strings.Split(value, " and ")
				for _, author := range authors {
					fmt.Fprintf(ris, "\n%s  - %s", "AU", author)
				}
			case "date":
				date := strings.Split(value, "-")
				fmt.Fprintf(ris, "\n%s  - %s", "Y1", date[0])
				if len(date) > 1 {
					fmt.Fprintf(ris, "/%s", date[1])
				}
				fmt.Fprint(ris, "//")
			case "urldate":
				date := strings.Split(value, "-")
				fmt.Fprintf(ris, "\n%s  - %s", "Y2", date[0])
				if len(date) > 1 {
					fmt.Fprintf(ris, "/%s", date[1])
				}
				fmt.Fprint(ris, "//")
			default:
				if f := ex.parse(field); f != "" {
					fmt.Fprintf(ris, "\n%s  - %s", f, value)
				}
			}
		}
	}

	// field  - value
	for field := range e.Optional {
		if value := e.Optional[field]; value != "" {
			switch field {
			case "editor":
				authors := strings.Split(value, " and ")
				for _, author := range authors {
					fmt.Fprintf(ris, "\n%s  - %s", "ED", author)
				}
			case "urldate":
				date := strings.Split(value, "-")
				fmt.Fprintf(ris, "\n%s  - %s", "Y2", date[0])
				if len(date) > 1 {
					fmt.Fprintf(ris, "/%s", date[1])
				}
				fmt.Fprint(ris, "//")
			case "pages":
				pages := strings.Split(value, "-")
				fmt.Fprintf(ris, "\n%s  - %s", "SP", pages[0])
				fmt.Fprintf(ris, "\n%s  - %s", "EP", pages[len(pages)-1])
			default:
				if f := ex.parse(field); f != "" {
					fmt.Fprintf(ris, "\n%s  - %s", f, value)
				}
			}
		}
	}

	ris.WriteString("\nER  - ")

	return ris.String()
}

var ris = &exRIS{
	dict: map[string]string{
		"online":        "ELEC",
		"article":       "JOUR",
		"thesis":        "THES",
		"inproceedings": "CPAPER",
		"book":          "BOOK",
		"inbook":        "CHAP",
		"patent":        "PAT",
		"report":        "RPRT",
		"title":         "TI",
		"journaltitle":  "JO",
		"booktitle":     "T2",
		"doi":           "DO",
		"number":        "M1",
		"abstract":      "N2",
		"publisher":     "PB",
		"ISBN":          "SN",
		"ISSN":          "SN",
		"url":           "UR",
		"volume":        "VL",
	},
}
