package main

import (
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
)

type Entry struct {
	Seq      int      `xml:"ent_seq"`
	Kanji    []Kanji  `xml:"k_ele"`
	Readings []string `xml:"r_ele>reb"`
	Senses   []Sense  `xml:"sense"`
}
type Kanji struct {
	Text      string   `xml:"keb"`
	Frequency []string `xml:"ke_pri"` // How often its seen this is defined as
	Info      string   `xml:"ke_inf"`
}
type Sense struct {
	Pos     []string `xml:"pos"`  // There are multiple possible poses v5r & vi == godan and intransitive
	Xrefs   []string `xml:"xref"` // References
	Ants    []string `xml:"ant"`  // Words opposite
	Glosses []string `xml:"gloss"`
}

func parseJMdict(path string, handle func(Entry) error) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	dec := xml.NewDecoder(gz)
	dec.Strict = false

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "entry" {
			continue
		}
		var e Entry
		if err := dec.DecodeElement(&e, &se); err != nil {
			return err
		}
		if err := handle(e); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	count := 0
	err := parseJMdict("JMdict_e.gz", func(e Entry) error {
		count++
		if count%20000 == 0 {
			fmt.Printf("%d: %+v\n", count, e.Kanji)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("total:", count)
}
