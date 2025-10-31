// Package handlers
package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

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
	url := fmt.Sprintf("https://httpbin.org/stream/%s", numChunks)
	res, err := http.Get(url)
	defer func() {
		err := res.Body.Close()
		if err != nil {
			server.WriteError(w, &server.HandlerError{
				Code:    500,
				Message: "Internal Server Error",
			}, err.Error())
			return
		}
	}()
	if err != nil {
		server.WriteError(w, &server.HandlerError{
			Code:    500,
			Message: "Internal Server Error",
		}, "Error getting response from httpbin")
		return
	}

	// Status line
	err = w.WriteStatusLine(response.OK)
	if err != nil {
		server.WriteError(w, &server.HandlerError{
			Code:    500,
			Message: "Internal Server Error",
		}, "Error writing status line")
	}

	// Headers
	hdrs := response.GetDefaultHeaders(0)
	hdrs.Delete("Content-Length")
	hdrs.Delete("Content-Type")
	hdrs.Add("Transfer-Encoding", "chunked")
	hdrs.Add("Content-Type", res.Header.Get("Content-Type"))
	err = w.WriteHeaders(hdrs)
	if err != nil {
		server.WriteError(w, &server.HandlerError{
			Code:    500,
			Message: "Internal Server Error",
		}, "Error writing headers")
	}

	// Preparing and sending response
	buf := make([]byte, 1024)
	for {
		n, err := res.Body.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			server.WriteError(w, &server.HandlerError{
				Code:    500,
				Message: "Internal Server Error",
			}, err.Error())
			return
		}
		if n > 0 {
			err = w.WriteChunkedBody(buf[:n])
			if err != nil {
				server.WriteError(w, &server.HandlerError{
					Code:    500,
					Message: "Internal Server Error",
				}, err.Error())
				return
			}
		}
		clear(buf)
	}
	_ = w.WriteChunkedBodyDone()
}
