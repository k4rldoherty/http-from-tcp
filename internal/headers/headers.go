// Package headers - to parse the field lines of a http request
package headers

import (
	"bytes"
	"fmt"
	"strings"
)

const CRLF = "\r\n"

type Headers map[string]string

func NewHeaders() Headers {
	return map[string]string{}
}

func (h Headers) Set(key, value string) {
	h[key] = value
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
	if key != strings.TrimRight(key, " ") {
		return 0, false, fmt.Errorf("invalid header name: %s", key)
	}

	value := bytes.TrimSpace(parts[1])
	key = strings.TrimSpace(key)
	h.Set(key, string(value))
	return rnIdx + 2, false, nil
}
