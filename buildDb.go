package main

import (
	"compress/gzip"
	"database/sql"
	"encoding/xml"
	"io"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

type Entry struct {
	ID       int      `xml:"ent_seq"`
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

func createDB() error {
	db, err := sql.Open("sqlite", "lookup.db")
	if err != nil {
		return err
	}
	defer db.Close()

	// PRimary key will be ent_seq from source
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS entries (id INTEGER PRIMARY KEY)`); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	count := 0
	err = parseJMdict("JMdict_e.gz", func(e Entry) error {
		if _, err := tx.Exec(`INSERT INTO entries(id) VALUES(?)`, e.ID); err != nil {
			return err
		}
		count++
		return nil
	})
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM entries`).Scan(&n); err != nil {
		return err
	}
	log.Printf("Rows in db = %d", n)

	return nil
}

func main() {
	if err := createDB(); err != nil {
		log.Fatal(err)
	}
}
