package jmdict

import (
	"database/sql"
	"log"
	"strings"

	_ "modernc.org/sqlite"
)

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

func cleanEntities(vals []string) {
	for i, v := range vals {
		vals[i] = strings.Trim(v, "&;")
	}
}

func CreateDB() error {
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
