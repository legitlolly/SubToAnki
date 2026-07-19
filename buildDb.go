package main

import (
	"compress/gzip"
	"database/sql"
	"encoding/xml"
	"io"
	"log"
	"os"
	"strings"

	_ "modernc.org/sqlite"
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

// DB SCHEMA
const schema = `
CREATE TABLE IF NOT EXISTS entries (
	id INTEGER PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS kanji (
	entry_id  INTEGER,
	text      TEXT,
	frequency TEXT,
	info      TEXT
);

CREATE TABLE IF NOT EXISTS readings (
	entry_id  INTEGER,
	text      TEXT,
	frequency TEXT,
	info      TEXT
);

CREATE TABLE IF NOT EXISTS senses (
	id       INTEGER PRIMARY KEY,
	entry_id INTEGER,
	pos      TEXT
);

CREATE TABLE IF NOT EXISTS glosses (
	sense_id INTEGER,
	text     TEXT
);
`

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

func cleanEntities(vals []string) {
	for i, v := range vals {
		vals[i] = strings.Trim(v, "&;")
	}
}

func createDB() error {
	db, err := sql.Open("sqlite", "lookup.db")
	if err != nil {
		return err
	}
	defer db.Close()

	if _, err := db.Exec(schema); err != nil {
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
		for _, k := range e.Kanji {
			cleanEntities(k.Info)
			if _, err := tx.Exec(
				`INSERT INTO kanji(entry_id, text, frequency, info) VALUES(?, ?, ?, ?)`,
				e.ID, k.Text, strings.Join(k.Frequency, ","), strings.Join(k.Info, ","),
			); err != nil {
				return err
			}
		}

		for _, r := range e.Readings {
			cleanEntities(r.Info)
			if _, err := tx.Exec(
				`INSERT INTO readings(entry_id, text, frequency, info) VALUES(?, ?, ?, ?)`,
				e.ID, r.Text, strings.Join(r.Frequency, ","), strings.Join(r.Info, ","),
			); err != nil {
				return err
			}
		}

		for _, s := range e.Senses {
			cleanEntities(s.Pos)
			res, err := tx.Exec(
				`INSERT INTO senses(entry_id, pos) VALUES(?, ?)`,
				e.ID, strings.Join(s.Pos, ","),
			)
			if err != nil {
				return err
			}
			senseID, err := res.LastInsertId()
			if err != nil {
				return err
			}
			for _, g := range s.Glosses {
				if _, err := tx.Exec(
					`INSERT INTO glosses(sense_id, text) VALUES(?, ?)`,
					senseID, g,
				); err != nil {
					return err
				}
			}
		}
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
	os.Remove("lookup.db")
	if err := createDB(); err != nil {
		log.Fatal(err)
	}
}
