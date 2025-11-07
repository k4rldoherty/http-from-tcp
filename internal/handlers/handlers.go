// Package handlers
package handlers

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/k4rldoherty/http-from-tcp/internal/headers"
	"github.com/k4rldoherty/http-from-tcp/internal/request"
	"github.com/k4rldoherty/http-from-tcp/internal/response"
	"github.com/k4rldoherty/http-from-tcp/internal/server"
)

var (
	serverErrorBody = []byte("<html><head><title>500 Internal Server Error</title></head><body><h1>Internal Server Error</h1><p>Okay, you know what? This one is on me.</p></body></html>")
	badRequestBody  = []byte("<html><head><title>400 Bad Request</title></head><body><h1>Bad Request</h1><p>Your request honestly kinda sucked.</p></body></html>")
	successBody     = []byte("<html><head><title>200 OK</title></head><body><h1>Success!</h1><p>Your request was an absolute banger.</p></body></html>")
)

func HandleYourProblem(w *response.Writer, r *request.Request) {
	err := &server.HandlerError{
		Code:    400,
		Message: string(badRequestBody),
	}
	server.WriteError(w, err, "Could not handle your request")
}

func HandleMyProblem(w *response.Writer, r *request.Request) {
	err := &server.HandlerError{
		Code:    500,
		Message: string(serverErrorBody),
	}
	server.WriteError(w, err, "Could not handle your request")
}

func HandleOther(w *response.Writer, r *request.Request) {
	if err := w.WriteStatusLine(response.OK); err != nil {
		log.Printf("ERROR: handle/WriteStatusLine: %v", err)
		server.WriteError(w, &server.HandlerError{
			Code:    500,
			Message: "Server error",
		}, "Could not write status line")
		return
	}

	// create headers
	hdrs := response.GetDefaultHeaders(len(successBody))
	hdrs.Set("Content-Type", "text/html")

	if err := w.WriteHeaders(hdrs); err != nil {
		log.Printf("Handle/WriteHeaders: %v", err)
		server.WriteError(w, &server.HandlerError{
			Code:    500,
			Message: "Server error",
		}, "Could not write headers")
		return
	}

	_, err := w.WriteBody(successBody)
	if err != nil {
		log.Printf("Handle: %v", err)
		server.WriteError(w, &server.HandlerError{
			Code:    500,
			Message: "Server error",
		}, "Could not write body")
		return
	}
}

func HandleHTTPBin(w *response.Writer, r *request.Request) {
	// Url and intitial get
	targetParts := strings.Split(r.RequestLine.Target, "/")
	numChunks := targetParts[len(targetParts)-1]
	url := fmt.Sprintf("https://httpbin.org/%s", numChunks)
	res, err := http.Get(url)
	defer func() {
		err := res.Body.Close()
		if err != nil {
			log.Printf("error closing response body: %v", err)
			return
		}
	}()
	if err != nil {
		log.Printf("error getting response from httpbin: %v", err)
		return
	}

	// Status line
	err = w.WriteStatusLine(response.OK)
	if err != nil {
		log.Printf("error writing status line: %v", err)
		return
	}

	// Headers
	hdrs := response.GetDefaultHeaders(0)
	hdrs.Delete("Content-Length")
	hdrs.Delete("Content-Type")
	hdrs.Add("Transfer-Encoding", "chunked")
	hdrs.Add("Content-Type", res.Header.Get("Content-Type"))
	hdrs.Add("Trailer", "X-Content-SHA256, X-Content-Length")
	err = w.WriteHeaders(hdrs)
	if err != nil {
		log.Printf("error writing headers: %v", err)
		return
	}

	rawBytes := 0
	bufForHash := []byte{}
	// Preparing and sending response
	buf := make([]byte, 1024)
	for {
		n, err := res.Body.Read(buf)
		bufForHash = append(bufForHash, buf[:n]...)
		rawBytes += n
		if err == io.EOF {
			break
		}
		if err != nil {
			return
		}
		if n > 0 {
			err = w.WriteChunkedBody(buf[:n])
			if err != nil {
				log.Printf("error writing chunked body: %v", err)
				return
			}
		}
		clear(buf)
	}
	_ = w.WriteChunkedBodyDone()
	trlrs := headers.Headers{}
	trlrs.Set("X-Content-Length", fmt.Sprintf("%v", rawBytes))
	hash := sha256.Sum256(bufForHash)
	trlrs.Set("X-Content-SHA256", fmt.Sprintf("%x", hash[:]))
	err = w.WriteTrailer(trlrs)
	if err != nil {
		log.Printf("error writing trailers: %v", err)
		return
	}
}

func HandleVideo(w *response.Writer, r *request.Request) {
	vid, err := os.ReadFile("assets/vim.mp4")
	if err != nil {
		log.Printf("error reading video file: %v", err)
		return
	}
	if err := w.WriteStatusLine(response.OK); err != nil {
		log.Printf("ERROR: handle/WriteStatusLine: %v", err)
		return
	}

	// create headers
	hdrs := response.GetDefaultHeaders(len(successBody))
	hdrs.Set("Content-Type", "video/mp4")

	if err := w.WriteHeaders(hdrs); err != nil {
		log.Printf("handle/WriteHeaders: %v", err)
		return
	}

	_, err = w.WriteBody(vid)
	if err != nil {
		log.Printf("Handle: %v", err)
	}
}
