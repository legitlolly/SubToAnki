package jmdict

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// bundleVersion is bumped whenever the on-disk shape of an entry line changes. Allowing for auto updates
const bundleVersion = 1

// Entries are emitted as positional arrays.
// [id, kanji[], readings[], senses[], priority, searchOnly[]]
// where each sense is [posCSV, glosses[]]. searchOnly holds the "sK"/"sk"
type exportEntry []any

type exportSense []any

func isSearchOnly(info []string) bool {
	for _, v := range info {
		if v == "sK" || v == "sk" {
			return true
		}
	}
	return false
}

// ExportBundle parses JMdict and writes a gzipped NDJSON bundle for the browser
// extension to import. The first line is a metadata object, every line after it
// is one entry.
func ExportBundle(srcPath, outPath string) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewWriterLevel(f, gzip.BestCompression)
	if err != nil {
		return err
	}

	w := bufio.NewWriterSize(gz, 1<<20)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)

	if err := enc.Encode(map[string]any{
		"format":  "subtoanki-jmdict-ndjson",
		"version": bundleVersion,
	}); err != nil {
		return err
	}

	count := 0
	err = parseJMdict(srcPath, func(e Entry) error {
		var kanji, readings, searchOnly []string
		priority := 999

		for _, k := range e.Kanji {
			cleanEntities(k.Info)
			if p := calculatePriority(k.Frequency); p < priority {
				priority = p
			}
			if isSearchOnly(k.Info) {
				searchOnly = append(searchOnly, k.Text)
				continue
			}
			kanji = append(kanji, k.Text)
		}

		for _, r := range e.Readings {
			cleanEntities(r.Info)
			if p := calculatePriority(r.Frequency); p < priority {
				priority = p
			}
			if isSearchOnly(r.Info) {
				searchOnly = append(searchOnly, r.Text)
				continue
			}
			readings = append(readings, r.Text)
		}

		senses := make([]exportSense, 0, len(e.Senses))
		for _, s := range e.Senses {
			cleanEntities(s.Pos)
			if len(s.Glosses) == 0 {
				continue
			}
			senses = append(senses, exportSense{strings.Join(s.Pos, ","), s.Glosses})
		}
		// Entries with no glosses are unusable
		if len(senses) == 0 {
			return nil
		}

		count++
		return enc.Encode(exportEntry{
			e.ID,
			orEmpty(kanji),
			orEmpty(readings),
			senses,
			priority,
			orEmpty(searchOnly),
		})
	})
	if err != nil {
		return err
	}

	if err := w.Flush(); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}

	st, err := f.Stat()
	if err != nil {
		return err
	}
	log.Printf("wrote %s: %d entries, %.1f MB gzipped", outPath, count, float64(st.Size())/1e6)
	return nil
}

func orEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
