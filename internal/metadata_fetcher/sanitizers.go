package metadata_fetcher

import (
	"bytes"
	"github.com/tailscale/hujson"
	"golang.org/x/text/runes"
	"regexp"
)

func sanitizeJSON(b []byte) []byte {
	b = sanitizeBytes(b)
	// remove escaped nul unicode byte sequences
	b = jsonInvalidRunes.ReplaceAll(b, invalidUTFSeq)
	// try remove redundant comas & comments
	hPayload, err := hujson.Parse(b)
	if err != nil {
		return b
	}

	hPayload.Standardize()
	return hPayload.Pack()
}

var jsonInvalidRunes = regexp.MustCompile(`(?m)\\u0000`)

var invalidUTFSeq = []byte("\ufffd")

func sanitizeBytes(b []byte) []byte {
	// remove invalid UTF8 escape seq
	b = bytes.ToValidUTF8(b, invalidUTFSeq)
	// remove redundant line endings
	b = bytes.TrimSpace(b)
	// remove nul unicode byte sequences - as they are not supported by postgres
	b = runes.Remove(runes.Predicate(func(r rune) bool {
		return r == rune(0)
	})).Bytes(b)
	return b
}
