type Entry struct {
	Seq      int      `xml:"ent_deq"`
	Kanji    []string `xml:"k_ele>keb"`
	Readings []string `xml:"r_ele>reb"`
	Senses   []Sense  `xml:"sense"`
}
type Sense struct {
	Pos     []string `xml:"pos"`  // There are multiple possible poses v5r & vi == godan and intransitive
	Xrefs   []string `xml:"xref"` // References
	Ants    []string `xml:"ant"`
	Glosses []string `xml:"gloss"`
}
