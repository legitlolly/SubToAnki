package jmdict

import (
	"database/sql"
	"fmt"
	"strings"
)

type senseGroup struct {
	pos     string
	glosses []string
}

func Lookup(word string) error {
	db, err := sql.Open("sqlite", "lookup.db")
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.Query(`
	SELECT entry_id FROM kanji WHERE text = ?
	UNION
	SELECT entry_id FROM readings WHERE text = ?
	`, word, word)
	if err != nil {
		return err
	}
	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if len(ids) == 0 {
		fmt.Println("not found")
		return nil
	}
	for _, id := range ids {
		var kanjiForms, readingForms []string

		krows, err := db.Query(`SELECT text, info FROM kanji WHERE entry_id = ?`, id)
		if err != nil {
			return err
		}
		for krows.Next() {
			var text, info string
			if err := krows.Scan(&text, &info); err != nil {
				krows.Close()
				return err
			}
			if info == "sK" {
				continue
			}
			kanjiForms = append(kanjiForms, text)
		}
		krows.Close()

		rrows, err := db.Query(`SELECT text FROM readings WHERE entry_id = ?`, id)
		if err != nil {
			return err
		}
		for rrows.Next() {
			var text string
			if err := rrows.Scan(&text); err != nil {
				rrows.Close()
				return err
			}
			readingForms = append(readingForms, text)
		}
		rrows.Close()

		head := strings.Join(kanjiForms, "、")
		if head == "" {
			head = strings.Join(readingForms, "、")
		} else {
			head += " 【" + strings.Join(readingForms, "、") + "】"
		}
		fmt.Printf("\n%s\n", head)

		srows, err := db.Query(`
			SELECT s.id, s.pos, g.text
			FROM senses s JOIN glosses g ON g.sense_id = s.id
			WHERE s.entry_id = ?
			ORDER BY s.id`, id)
		if err != nil {
			return err
		}

		type senseGroup struct {
			pos     string
			glosses []string
		}
		var groups []senseGroup
		lastID := -1
		for srows.Next() {
			var sid int
			var pos, gloss string
			if err := srows.Scan(&sid, &pos, &gloss); err != nil {
				srows.Close()
				return err
			}
			if sid != lastID {
				groups = append(groups, senseGroup{pos: pos})
				lastID = sid
			}
			g := &groups[len(groups)-1]
			g.glosses = append(g.glosses, gloss)
		}
		srows.Close()

		for i, s := range groups {
			fmt.Printf("  %d. [%s] %s\n", i+1, s.pos, strings.Join(s.glosses, "; "))
		}
	}
	return nil
}
