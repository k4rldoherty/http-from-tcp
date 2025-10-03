// Package headers - to parse the field lines of a http request
package headers

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

const CRLF = "\r\n"

type Headers map[string]string

func NewHeaders() Headers {
	return map[string]string{}
}

func (h Headers) Set(key, value string) {
	h[key] = value
}

func (h Headers) Get(key string) string {
	if v, ok := h[key]; ok {
		return v
	}
	return ""
}

func (h Headers) Parse(d []byte) (n int, done bool, err error) {
	rnIdx := bytes.Index(d, []byte(CRLF))
	if rnIdx == -1 {
		return 0, false, nil
	}
	if rnIdx == 0 {
		return 2, true, nil
	}
	parts := bytes.SplitN(d[:rnIdx], []byte(":"), 2)
	key := string(parts[0])
	key = strings.ToLower(key)
	if key != strings.TrimRight(key, " ") {
		return 0, false, fmt.Errorf("invalid header name: %s", key)
	}
	key = strings.TrimSpace(key)
	if !isValidKey(key) {
		return 0, false, fmt.Errorf("invalid key formatting: %s", key)
	}
	value := string(bytes.TrimSpace(parts[1]))
	if v, ok := h[key]; ok {
		joinedV := v + ", " + value
		h.Set(key, joinedV)
	} else {
		h.Set(key, value)
	}
	return rnIdx + 2, false, nil
}

func isValidKey(key string) bool {
	if len(key) < 1 {
		return false
	}
	for _, l := range key {
		if unicode.IsDigit(l) ||
			unicode.IsLetter(l) ||
			l == rune('!') ||
			l == rune('#') ||
			l == rune('$') ||
			l == rune('%') ||
			l == rune('`') ||
			l == rune('*') ||
			l == rune('+') ||
			l == rune('-') ||
			l == rune('.') ||
			l == rune('^') ||
			l == rune('_') ||
			l == rune('|') ||
			l == rune('~') ||
			l == rune('&') {
			continue
		} else {
			return false
		}
	}
	return true
}
