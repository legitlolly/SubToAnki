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

func main() {
	data, err := os.Open("JMdict_e.gz")
	if err != nil {
		log.Fatal(err)
	}
	defer data.Close()

	gz, err := gzip.NewReader(data)
	if err != nil {
		log.Fatal(err)
	}
	defer gz.Close()

	dec := xml.NewDecoder(gz)
	dec.Strict = false

	count := 0
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "entry" {
			continue
		}

		var e Entry
		if err := dec.DecodeElement(&e, &se); err != nil {
			log.Fatal(err)
		}
		count++
		if count%20000 == 0 {
			fmt.Printf("%d: %+v\n", count, e.Kanji)
		}
	}
	fmt.Println("total entries:", count)
}
