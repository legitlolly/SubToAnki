package main

import (
	"flag"
	"log"

	"subtoanki/internal/jmdict"
)

func main() {
	src := flag.String("src", "JMdict_e.gz", "path to JMdict_e.gz")
	out := flag.String("out", "extension/dict/jmdict.ndjson.gz", "output bundle path")
	flag.Parse()

	if err := jmdict.ExportBundle(*src, *out); err != nil {
		log.Fatal(err)
	}
}
