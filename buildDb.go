package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
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
	data, err := os.ReadFile("example.xml")
	if err != nil {
		log.Fatal(err)
	}

	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.Strict = false

	var e Entry
	if err := dec.Decode(&e); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", e)
}
