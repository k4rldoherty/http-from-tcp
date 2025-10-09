// Package request
package request

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/k4rldoherty/tcp-from-http/internal/headers"
)

const BUFFERSIZE = 8

type ParserState int

const (
	Initialized ParserState = iota
	ParsingHeaders
	ParsingBody
	Done
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	State       ParserState
}

type RequestLine struct {
	Method      string
	Target      string
	HTTPVersion string
}

func (r *Request) Parse(data []byte, hitEOF bool) (int, error) {
	totalBytesParsed := 0
	for r.State != Done {
		n, err := r.parseSingle(data[totalBytesParsed:], hitEOF)
		if err != nil {
			return 0, err
		}
		totalBytesParsed += n
		if n == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte, hitEOF bool) (int, error) {
	switch r.State {
	case Initialized:
		bytes, rl, err := parseRequestLine(string(data))
		if err != nil {
			return 0, err
		}
		if bytes == 0 {
			return 0, nil
		}
		r.RequestLine = *rl
		r.State = ParsingHeaders
		return bytes, nil
	case ParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.State = ParsingBody
		}
		return n, nil
	case ParsingBody:
		cl := r.Headers.Get("content-length")
		if cl == "" {
			r.State = Done
			return 0, nil
		}
		i, err := strconv.Atoi(cl)
		if err != nil {
			return 0, err
		}
		r.Body = append(r.Body, data...)
		bodyLen := len(r.Body)
		if bodyLen == i {
			r.State = Done
		} else if bodyLen > i {
			return 0, errors.New("body longer than content length")
		} else {
			if hitEOF {
				return 0, errors.New("body shorter than content length")
			}
		}
		return len(data), nil
	case Done:
		return 0, errors.New("parse function called in Done state")
	default:
		return 0, errors.New("unknown state")
	}
}

func RequestFromReader(r io.Reader) (*Request, error) {
	// buffer to read data into
	buf := make([]byte, BUFFERSIZE)
	// how much data we have read
	// from the reader into the buffer
	readToIndex := 0
	// initialize request with state
	// as initialized
	request := Request{
		State:   Initialized,
		Headers: headers.NewHeaders(),
		Body:    []byte{},
	}
	hitEOF := false
	for request.State != Done {
		// if the buffer is full
		// create a new one twice the size and copy the data in
		if readToIndex == len(buf) {
			nb := make([]byte, len(buf)*2)
			copy(nb, buf[:readToIndex])
			buf = nb
		}
		// if end of file flag is false
		// read into the buffer from the prev readToIndex
		// check if EOF
		// and if not increase readToIndex by the bytes read into the buffer
		if !hitEOF {
			n, err := r.Read(buf[readToIndex:])
			if errors.Is(err, io.EOF) {
				hitEOF = true
			} else if err != nil {
				return nil, err
			} else if n > 0 {
				readToIndex += n
			}
		}
		// read from r into buffer starting at readToIndex
		// parse read in bytes
		consumed, err := request.Parse(buf[:readToIndex], hitEOF)
		if err != nil {
			return nil, err
		}
		// remove parsed data from buffer
		if consumed > 0 {
			copy(buf, buf[consumed:readToIndex])
			// decrement readToIndex by the bytes parsed
			readToIndex -= consumed
		}

		if hitEOF && readToIndex == 0 {
			break
		}
	}
	return &request, nil
}

func parseRequestLine(rl string) (int, *RequestLine, error) {
	rlEnd := strings.Index(rl, "\r\n")
	if rlEnd == -1 {
		return 0, nil, nil
	}
	rl = rl[:rlEnd]
	bytesConsumed := rlEnd + 2
	rlParts := strings.Split(rl, " ")
	if len(rlParts) != 3 {
		return 0, nil, fmt.Errorf("incorrect number of sections of request line: %d", len(rlParts))
	}
	method, err := parseMethod(rlParts[0])
	if err != nil {
		return 0, nil, err
	}
	target, err := parseTarget(rlParts[1])
	if err != nil {
		return 0, nil, err
	}
	version, err := parseVersion(rlParts[2])
	if err != nil {
		return 0, nil, err
	}
	reqL := RequestLine{
		Method:      method,
		Target:      target,
		HTTPVersion: version,
	}
	return bytesConsumed, &reqL, nil
}

func parseMethod(method string) (string, error) {
	match, err := regexp.MatchString("^[A-Z]+", method)
	if !match || err != nil {
		return "", fmt.Errorf("invalid method %s", method)
	}
	return method, nil
}

func parseTarget(target string) (string, error) {
	return target, nil
}

func parseVersion(version string) (string, error) {
	vp := strings.Split(version, "/")
	if len(vp) != 2 {
		return "", fmt.Errorf("invalid version format: %s", version)
	}
	v := vp[1]
	if v != "1.1" {
		return "", fmt.Errorf("invalid version: %s. http version 1.1 is supported", vp[1])
	}
	return vp[1], nil
}
