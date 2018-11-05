# Scholar Library
Go library to generate and manipulate bibliography entries.

For the CLI program, click [here](https://github.com/cgxeiji/scholar/).

## Usage

Scholar uses a YAML configuration file to load the types of entries to be used.
The basic structure of the file is as follows:

```yaml
typeName:
  desc: A brief description of the type.
  req:
    field1: A brief description of field1.
    field2: A brief description of field2.
    field3: A brief description of field3.
  opt:
    field1: A brief description of field1.
    field2: A brief description of field2.
    field3: A brief description of field3.
```

For example:
```yaml
article:
  desc: An article in a journal, magazine, newspaper, or other periodical which forms a self-contained unit with its own title.
  req:
    author: Author(s) of the article.
    title: Title of the article.
    journaltitle: Title of the journal.
    date: YYYY-MM-DD format.
  opt:
    editor: Editor(s) of the journal.
    language: Language of the article.
    series: Series of the journal.
    volume: Volume of the journal.
    number: Number of the journal.
    issn: ISSN number of the article.
    doi: DOI code of the article.
    url: URL of the article.
    urldate: Access date in YYY-MM-DD format.

book:
  desc: A single-volume book with one or more authors where the authors share credit for the work as a whole.
  req:
    author: Author(s) of the book.
    title: Title of the book.
    date: YYYY-MM-DD format.
  opt:
    editor: Editor(s) of the book.
    publisher: Publisher of the book.
    location: Location of the publisher.
    language: Language of the book.
    series: Series of the book.
    volume: Volume of the book.
    number: Number of the book.
    pages: Number of pages is the book.
    isbn: ISBN number of the book.
    doi: DOI code of the book.
    url: URL of the book.
    urldate: Access date in YYY-MM-DD format.
```

To load the entry types to be used by Scholar, do:

```go
func main() {
    if err := scholar.LoadTypes("types.yaml"); err != nil {
        // handle error
    }    
}
```

To create a new Entry struct, do:
```go
    entry := scholar.NewEntry("article")
```

To change the type of an entry, do:
```go
    entryArticle := scholar.NewEntry("article")
    entryBook := scholar.Convert(entryArticle, "book")
```

To get a BibLaTex format reference, do:
```go
    entry := scholar.NewEntry("article")
    entry.Required["title"] = "The Article"
    entry.Required["author'] = "Last, First and Other, Second"
    entry.Required["journaltitle"] = "The Journal of Articles"
    entry.Required["date"] = "2018"

    fmt.Println(entry.Bib())
```
```
output:
    @article{last2018,
      author = {Last, First and Other, Second},
      date = {2018},
      journaltitle = {The Journal of Articles},
      title = {The Article}
    }
```


## TODO

- [ ] Return an error on `TypeNotFound` for `NewEntry()`
- [ ] Improve documentation
