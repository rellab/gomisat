#!/bin/sh
# Download the benchmark corpus described by corpus/manifest.tsv.
#
#   scripts/fetch-corpus.sh [destination]
#
# The destination defaults to $GOMISAT_CORPUS, then to ~/.cache/gomisat/corpus.
# It is deliberately outside the repository: the instances are large, they are
# reproducible from the manifest, and the repository lives in a synced folder.
#
# Each archive is extracted into <destination>/<local-dir>, flattened, so the
# expected answer is carried by the directory name. Archives already present are
# not downloaded again. With a checksum column in the manifest the download is
# verified against it and a mismatch aborts the run; otherwise the computed
# checksum is printed so it can be pinned.
set -eu

base=https://www.cs.ubc.ca/~hoos/SATLIB
root=$(cd "$(dirname "$0")/.." && pwd)
manifest=$root/corpus/manifest.tsv
dest=${1:-${GOMISAT_CORPUS:-$HOME/.cache/gomisat/corpus}}
cache=$dest/.archives

mkdir -p "$cache"
printf 'corpus: %s\n' "$dest"

sed -e 's/#.*$//' -e '/^[[:space:]]*$/d' "$manifest" | while IFS="$(printf '\t')" read -r dir url expected sha note; do
	[ -n "${dir:-}" ] || continue
	archive=$cache/$(basename "$url")
	if [ ! -f "$archive" ]; then
		printf 'fetching %s\n' "$url"
		curl -fsSL --retry 3 -o "$archive.part" "$base/$url"
		mv "$archive.part" "$archive"
	fi
	sum=$(shasum -a 256 "$archive" | cut -d' ' -f1)
	if [ -n "${sha:-}" ] && [ "$sha" != "$sum" ]; then
		printf 'checksum mismatch for %s\n  manifest %s\n  download %s\n' "$archive" "$sha" "$sum" >&2
		exit 1
	fi

	if [ -d "$dest/$dir" ] && [ -n "$(find "$dest/$dir" -name '*.cnf' -print -quit)" ]; then
		count=$(find "$dest/$dir" -name '*.cnf' | wc -l | tr -d ' ')
		printf '%-26s %5s instances  (present)\n' "$dir" "$count"
		continue
	fi

	tmp=$(mktemp -d)
	tar xzf "$archive" -C "$tmp"
	mkdir -p "$dest/$dir"
	# Flatten: the archives disagree about how deeply they nest, and a few ship
	# the instances gzipped individually.
	find "$tmp" -name '*.cnf.gz' -exec gunzip {} +
	find "$tmp" -name '*.cnf' -exec mv {} "$dest/$dir/" \;
	rm -rf "$tmp"
	count=$(find "$dest/$dir" -name '*.cnf' | wc -l | tr -d ' ')
	printf '%-26s %5s instances  %s  sha256=%s\n' "$dir" "$count" "$expected" "$sum"
done

total=$(find "$dest" -name '*.cnf' | wc -l | tr -d ' ')
printf 'total %s instances in %s\n' "$total" "$dest"
