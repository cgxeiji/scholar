# Scholar: a CLI Reference Manager

Scholar is a terminal based reference manager written in Go, that helps you
keep track of your resources.  It uses YAML files to save the metadata. Entry
types and fields are taken from the [BibLaTex
format](http://mirrors.ctan.org/macros/latex/contrib/biblatex/doc/biblatex.pdf).
It was inspired by [papis](https://github.com/papis/papis), a CLI reference
manager written in python.

![Scholar
Demo](https://github.com/cgxeiji/scholar/raw/master/img/scholar_demo.gif)

## Features

Add any file using:
```
$ scholar add filename.ext
```

Specify DOI or search for metadata on the web (currently only support for CrossRef):
```
$ scholar add filename.ext --doi=10.1007/978-94-011-6022-3_3

$ scholar add general theory of relativity einstein 1992
```

Use multiple libraries for different types of entries:
```
$ scholar open einstein --library=research

$ scholar add Harry Potter.epub --library=books
```

Export entries to BibLaTex by default:
```
$ scholar export > references.bib

$ scholar export --library=research > research.bib
```

Or choose other formats:
```
$ scholar export --format=bibtex > references.bib

$ scholar export --format=ris > references.bib
```

And much more:
```
$ scholar help

Scholar: a CLI Reference Manager

Scholar is a CLI reference manager that keeps track of
your documents metadata using YAML files with biblatex format.

Usage:
  scholar [flags]
  scholar [command]

Available Commands:
  add         Add a new entry
  config      Configure Scholar
  edit        Edit an entry
  export      Export entries
  fetch       Prints the file path of the entry
  help        Help about any command
  import      Import a bibtex/biblatex file
  open        Open an entry
  remove      Remove an entry

Flags:
  -h, --help             help for scholar
  -i, --interactive      toggle interactive mode (enabled by default)
  -l, --library string   specify the library

Use "scholar [command] --help" for more information about a command.
```

## Installation

If you do not have Go installed, follow this guide:

- [The Go Programming Language: Getting Started](https://golang.org/doc/install)

Then, run:
```
go get -u github.com/cgxeiji/scholar
```

Done!

## Interactive Mode

By default, Scholar will launch a selection screen when `edit`, `open`,
`fetch`, or `remove` commands have a query with more than one entry.

If you want to use Scholar inside a script, you can disable interactive mode by
passing the flag `-i`, or setting `interactive: false` in the configuration
file.

When interactive mode is disabled, Scholar will return an `exit status 1` and
the number of entries found if there is more than one entry that matches the
query. Also, `remove` will delete the entry without confirmation.

## TODO

### General

- [x] Add `-i` flag to enable/disable interactive mode.
- [x] Add `interactive: true` settings in the configuration file.
- [x] Make attached file path relative to entry, ~unless is an external file~.
- [x] Be able to copy the file path to stdout.
- [ ] Be able to reference a file instead of copying it.
- [ ] Add support for attaching multiple files.

### Add

- [x] Add a flag for manual/auto input of metadata.
- [x] Handle non-interactive mode for add.

### Config

- [ ] Add function to create a local configuration file.

### Export

- [x] Add different export formats.

### Open

- [ ] Add selection menu if multiple files are attached.
- [ ] Open metadata if no file/url/DOI is attached.

### Remove

- [x] Add remove confirmation.

## License

MIT
