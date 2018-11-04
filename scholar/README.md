# Scholar: a CLI Reference Manager

Scholar is a terminal based reference manager written in Go, that helps you
keep track of your resources.  It uses YAML files to save the metadata. Entry
types and fields are taken from the [BibLaTex
format](http://mirrors.ctan.org/macros/latex/contrib/biblatex/doc/biblatex.pdf).
It was inspired by [papis](https://github.com/papis/papis), a CLI reference
manager written in python.

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
  help        Help about any command
  import      Import a bibtex/biblatex file
  open        Open an entry
  remove      Remove an entry

Flags:
  -h, --help             help for scholar
  -l, --library string   specify the library

Use "scholar [command] --help" for more information about a command.
```

## Installation

If you do not have Go installed, follow this guide:

- [The Go Programming Language: Getting Started](https://golang.org/doc/install)

Then, run:
```
go get -u github.com/cgxeiji/scholar/scholar
```

Done!

## TODO

### General

- [ ] Add `-i` flag to enable/disable interactive mode.
- [ ] Add `interactive: true` settings in the configuration file.
- [ ] Make attached file path relative to entry, unless is an external file.
- [ ] Be able to reference a file instead of copying it.
- [ ] Add support for attaching multiple files.

### Add

- [ ] Add a flag for manual/auto input of metadata.

### Config

- [ ] Add function to create a local configuration file.

### Export

- [ ] Add different export formats.

### Open

- [ ] Add selection menu if multiple files are attached.
- [ ] Open metadata if no file/url/DOI is attached.

### Remove

- [ ] Add remove confirmation.

## License

MIT
