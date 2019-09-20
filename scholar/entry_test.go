package scholar

import (
	"bytes"
	"fmt"
	"testing"
)

var mockEntryTypes = []byte(`# Mock Entry Types

article:
  desc: An article in a journal, magazine, newspaper, or other periodical which forms a self-contained unit with its own title.
  req:
    author: Author(s) of the article.
    title: Title of the article.
    journaltitle: Title of the journal.
    date: YYYY-MM-DD format.
    obscure: Obscure field for testing purposes.
  opt:
    doi: DOI code of the article.

book:
  desc: A single-volume book with one or more authors where the authors share credit for the work as a whole.
  req:
    author: Author(s) of the book.
    title: Title of the book.
    date: YYYY-MM-DD format.
  opt:
    editor: Editor(s) of the book.
    isbn: ISBN number of the book.
    publisher: Publisher of the book.
    doi: DOI code of the book.

misc:
  desc: A fallback for entries which do not fit into any other category.
  req:
    author: Author(s) of the work.
    title: Title of the work.
    date: YYYY-MM-DD format.
  opt:
    url: URL of the work.
    urldate: Access date in YYYY-MM-DD format.
`)

func TestLoadTypes(t *testing.T) {
	err := loadTypes(mockEntryTypes)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"article",
		"book",
		"misc",
	}

	for _, et := range want {
		t.Run(et, func(t *testing.T) {
			_, ok := EntryTypes[et]
			if !ok {
				t.Fatalf("could not find entry type %q", et)
			}
		})
	}
}

func TestTypesInfo(t *testing.T) {
	err := loadTypes(mockEntryTypes)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		t.Run(fmt.Sprintf("level %d", i), func(t *testing.T) {
			info := typesInfo(i)
			if info == "" {
				t.Errorf("typesInfo(%d) did not return any info", i)
			}
		})
	}
}

func TestFTypesInfo(t *testing.T) {
	err := loadTypes(mockEntryTypes)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		t.Run(fmt.Sprintf("level %d", i), func(t *testing.T) {
			b := new(bytes.Buffer)
			FTypesInfo(b, i)
			if b.Len() == 0 {
				t.Errorf("FTypesInfo(%d) did not return any info", i)
			}
		})
	}
}

func BenchmarkTypesInfo(b *testing.B) {
	err := loadTypes(mockEntryTypes)
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		typesInfo(3)
	}
}

func TestNewEntry(t *testing.T) {
	err := loadTypes(mockEntryTypes)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"article",
		"book",
		"misc",
	}

	for _, name := range want {
		t.Run(name, func(t *testing.T) {
			_, err = NewEntry(name)
			if err != nil {
				t.Fatalf("could not find entry type %q, got error:\n\t%v", name, err)
			}
		})
	}

	t.Run("test ErrTypeNotFound", func(t *testing.T) {
		_, err = NewEntry("test")
		if !IsError(ErrTypeNotFound, err) {
			t.Fatal("error other than ErrTypeNotFound:", err)
		}
		t.Log("Expected error:\n", err)
	})
}

func TestEntryType_Get(t *testing.T) {
	err := loadTypes(mockEntryTypes)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"article",
		"book",
		"misc",
	}

	for _, name := range want {
		et, ok := EntryTypes[name]
		if !ok {
			t.Errorf("EntryTypes does not have type %q\n\tEntryTypes = %v", name, EntryTypes)
			continue
		}

		entry := et.get()

		if entry.Type != name {
			t.Errorf("entry.Type does not match: got %q, want %q", entry.Type, name)
		}
	}
}

func BenchmarkEntryType_Get(b *testing.B) {
	err := loadTypes(mockEntryTypes)
	if err != nil {
		b.Fatal(err)
	}

	want := []string{
		"article",
		"book",
		"misc",
	}

	for i := 0; i < b.N; i++ {
		for _, name := range want {
			et, ok := EntryTypes[name]
			if !ok {
				b.Errorf("EntryTypes does not have type %q\n\tEntryTypes = %v", name, EntryTypes)
				continue
			}

			entry := et.get()

			if entry.Type != name {
				b.Errorf("entry.Type does not match: got %q, want %q", entry.Type, name)
			}
		}

	}
}

func TestConvert(t *testing.T) {
	err := loadTypes(mockEntryTypes)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"article",
		"book",
		"misc",
	}

	convertTo := "article"

	for _, name := range want {
		t.Run(name+" to "+convertTo, func(t *testing.T) {
			entry, err := NewEntry(name)
			if err != nil {
				t.Fatal(err)
			}

			entry, err = Convert(entry, convertTo)
			if IsError(ErrFieldNotFound, err) {
				t.Log("Expected error:\n", err)
			} else if err != nil {
				t.Fatal(err)
			}

			if entry.Type != convertTo {
				t.Errorf("converted entry.Type does not match: got %q, want %q", entry.Type, name)
			}
		})
	}

	t.Run("article to misc to article", func(t *testing.T) {
		article, err := NewEntry("article")
		if err != nil {
			t.Fatal(err)
		}

		article.Required["title"] = "The Title"
		article.Required["journaltitle"] = "The Journal"
		article.Required["date"] = "2006-01-02"
		article.Required["author"] = "Last, First and Other, Name"
		article.Required["obscure"] = "Testing Unicode: 𐌼𐌰𐌲 𐌲𐌻𐌴𐍃 𐌹̈𐍄𐌰𐌽, 𐌽𐌹 𐌼𐌹𐍃 𐍅𐌿 𐌽𐌳𐌰𐌽 𐌱𐍂𐌹𐌲𐌲𐌹𐌸"
		article.Optional["doi"] = "123/456789"

		misc, err := Convert(article, "misc")
		if err != nil {
			t.Fatal(err)
		}

		conv, err := Convert(misc, "article")
		if err != nil {
			t.Fatal(err)
		}

		if conv.Type != article.Type {
			t.Errorf("converted entry.Type does not match: got %q, want %q", conv.Type, article.Type)
		}
		for field, value := range conv.Required {
			if value != article.Required[field] {
				t.Errorf("converted entry.Required[%s] does not match: got %q, want %q", field, value, article.Required[field])
			}
		}
		for field, value := range conv.Optional {
			if value != article.Optional[field] {
				t.Errorf("converted entry.Optional[%s] does not match: got %q, want %q", field, value, article.Optional[field])
			}
		}
	})
}

func mockEntry() (*Entry, error) {
	err := loadTypes(mockEntryTypes)
	if err != nil {
		return nil, err
	}

	entry, err := NewEntry("article")
	if err != nil {
		return nil, err
	}

	entry.Required["title"] = "The Title"
	entry.Required["journaltitle"] = "The Journal"
	entry.Required["date"] = "2006-01-02"
	entry.Required["author"] = "Last, First and Other, Name"
	entry.Required["obscure"] = "Testing Unicode: 𐌼𐌰𐌲 𐌲𐌻𐌴𐍃 𐌹̈𐍄𐌰𐌽, 𐌽𐌹 𐌼𐌹𐍃 𐍅𐌿 𐌽𐌳𐌰𐌽 𐌱𐍂𐌹𐌲𐌲𐌹𐌸"
	entry.Optional["doi"] = "123/456789"

	return entry, nil
}

const mockBibOutput = `@article{last2006,
  author = {Last, First and Other, Name},
  date = {2006-01-02},
  journaltitle = {The Journal},
  obscure = {Testing Unicode: 𐌼𐌰𐌲 𐌲𐌻𐌴𐍃 𐌹̈𐍄𐌰𐌽, 𐌽𐌹 𐌼𐌹𐍃 𐍅𐌿 𐌽𐌳𐌰𐌽 𐌱𐍂𐌹𐌲𐌲𐌹𐌸},
  title = {The Title},
  doi = {123/456789}
}`

func TestEntry_Bib(t *testing.T) {
	entry, err := mockEntry()
	if err != nil {
		t.Fatal(err)
	}

	want := mockBibOutput
	got := entry.Bib()

	if got != want {
		t.Errorf("Bib() did not output the expected format:\ngot:\n%v\nwant:\n%v", got, want)
	}
}

func BenchmarkEntry_Bib(b *testing.B) {
	entry, err := mockEntry()
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		entry.Bib()
	}
}

func TestEntry_BibT(t *testing.T) {
	entry, err := mockEntry()
	if err != nil {
		t.Fatal(err)
	}

	want := mockBibOutput
	got := entry.bibT()

	if got != want {
		t.Errorf("bibT() did not output the expected format:\ngot:\n%v\nwant:\n%v", got, want)
	}
}

func BenchmarkEntry_bibT(b *testing.B) {
	entry, err := mockEntry()
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		entry.bibT()
	}
}
