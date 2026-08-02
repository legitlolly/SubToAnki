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

// JMdict marks rare or irregular forms "sK" (search-only kanji) and "sk"
// (search-only kana). They should still match a lookup, but must never be shown
// as a headword -- otherwise 大丈夫 lists だいじょーぶ alongside its real
// readings. info is stored comma-joined and a form can carry several tags at
// once ("ateji,sK"), so this is a membership test, not a string compare.
func searchOnly(info string) bool {
	for _, tag := range strings.Split(info, ",") {
		if tag == "sK" || tag == "sk" {
			return true
		}
	}
	return false
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

		krows, err := db.Query(`SELECT text, info FROM kanji WHERE entry_id = ? ORDER BY priority, rowid`, id)
		if err != nil {
			return err
		}
		for krows.Next() {
			var text, info string
			if err := krows.Scan(&text, &info); err != nil {
				krows.Close()
				return err
			}
			if searchOnly(info) {
				continue
			}
			kanjiForms = append(kanjiForms, text)
		}
		krows.Close()

		rrows, err := db.Query(`SELECT text, info FROM readings WHERE entry_id = ? ORDER BY priority, rowid`, id)
		if err != nil {
			return err
		}
		for rrows.Next() {
			var text, info string
			if err := rrows.Scan(&text, &info); err != nil {
				rrows.Close()
				return err
			}
			if searchOnly(info) {
				continue
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
