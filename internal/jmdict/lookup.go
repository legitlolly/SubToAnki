package jmdict

import (
	"database/sql"
	"fmt"
)

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
		rows, err := db.Query(`SELECT text, info FROM kanji WHERE entry_id = ?`, id)
		if err != nil {
			return err
		}
		for rows.Next() {
			var text, info string
			if err := rows.Scan(&text, &info); err != nil {
				rows.Close()
				return err
			}
			fmt.Println(id, text, info)
		}
		rows.Close()
	}
	return nil
}
