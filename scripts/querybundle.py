#!/usr/bin/env python3
"""Query the extension's dictionary bundle from the command line.

Deliberately prints in the same format as `go run ./cmd/subtoanki lookup <word>`
so the two can be diffed directly:

    go run ./cmd/subtoanki lookup 食べる > /tmp/go.txt
    ./scripts/querybundle.py 食べる      > /tmp/bundle.txt
    diff /tmp/go.txt /tmp/bundle.txt

A clean diff means the export dropped nothing the SQLite path keeps, which is
the only thing worth checking here -- the bundle is a build artifact, so any
disagreement is an export bug rather than a data bug.
"""

import argparse
import gzip
import json
import sys

DEFAULT_BUNDLE = "extension/dict/jmdict.ndjson.gz"


def entries(path):
    with gzip.open(path, "rt", encoding="utf-8") as f:
        header = json.loads(f.readline())
        if header.get("format") != "subtoanki-jmdict-ndjson":
            sys.exit(f"{path}: not a subtoanki bundle (got {header!r})")
        for line in f:
            line = line.strip()
            if line:
                yield json.loads(line)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("word", help="kanji or kana form to look up")
    ap.add_argument("--bundle", default=DEFAULT_BUNDLE)
    ap.add_argument(
        "--prefix",
        action="store_true",
        help="match any term starting with WORD instead of requiring an exact match",
    )
    ap.add_argument(
        "--verbose",
        action="store_true",
        help="also show id, priority and search-only forms (breaks diffing)",
    )
    ap.add_argument("--json", action="store_true", help="emit raw bundle lines")
    ap.add_argument("--limit", type=int, default=0, help="stop after N matches")
    args = ap.parse_args()

    found = 0
    for entry in entries(args.bundle):
        eid, kanji, readings, senses, priority, search_only = entry
        terms = kanji + readings + search_only

        if args.prefix:
            hit = any(t.startswith(args.word) for t in terms)
        else:
            hit = args.word in terms
        if not hit:
            continue

        found += 1
        if args.json:
            print(json.dumps(entry, ensure_ascii=False))
        else:
            head = "、".join(kanji)
            if head:
                head += " 【" + "、".join(readings) + "】"
            else:
                head = "、".join(readings)
            print(f"\n{head}")
            for i, (pos, glosses) in enumerate(senses, 1):
                print(f"  {i}. [{pos}] {'; '.join(glosses)}")
            if args.verbose:
                print(f"     -- id={eid} priority={priority} search_only={search_only}")

        if args.limit and found >= args.limit:
            break

    if not found:
        print("not found")


if __name__ == "__main__":
    main()
