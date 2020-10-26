package scholar

type exporter interface {
	export(*Entry) string
}

func getExporter(format string) exporter {
	switch format {
	case "bibtex":
		return bibtex
	case "biblatex":
		return biblatex
	case "ris":
		return ris
	}
	return biblatex
}
