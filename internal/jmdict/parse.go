package jmdict

import (
	"compress/gzip"
	"encoding/xml"
	"io"
	"os"
)

type Entry struct {
	ID       int       `xml:"ent_seq"`
	Kanji    []Kanji   `xml:"k_ele"`
	Readings []Reading `xml:"r_ele"`
	Senses   []Sense   `xml:"sense"`
}
type Kanji struct {
	Text      string   `xml:"keb"`
	Frequency []string `xml:"ke_pri"` // How often its seen this is defined as
	Info      []string `xml:"ke_inf"`
}
type Sense struct {
	Pos     []string `xml:"pos"`  // There are multiple possible poses v5r & vi == godan and intransitive
	Xrefs   []string `xml:"xref"` // References
	Ants    []string `xml:"ant"`  // Words opposite
	Glosses []string `xml:"gloss"`
}
type Reading struct {
	Text      string   `xml:"reb"`
	Frequency []string `xml:"re_pri"`
	Info      []string `xml:"re_inf"`
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
