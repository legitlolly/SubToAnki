package main

import (
	"log"
	"os"

	"subtoanki/internal/jmdict"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: subtoanki <ingest|lookup> [args]")
	}
	switch os.Args[1] {
	case "ingest":
		if err := jmdict.CreateDB(); err != nil {
			log.Fatal(err)
		}
	case "lookup":
		if len(os.Args) < 3 {
			log.Fatal("usage: subtoanki lookup <word_to_lookup>")
		}
		if err := jmdict.Lookup(); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("Unexpected command %q", os.Args[1])
	}
}
