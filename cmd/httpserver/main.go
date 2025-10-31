package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/k4rldoherty/http-from-tcp/internal/handlers"
	"github.com/k4rldoherty/http-from-tcp/internal/request"
	"github.com/k4rldoherty/http-from-tcp/internal/response"
	"github.com/k4rldoherty/http-from-tcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer func() {
		err := server.Close()
		if err != nil {
			log.Fatalf("Error closing server: %v", err)
		}
	}()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, r *request.Request) {
	url := r.RequestLine.Target
	if url == "/yourproblem" {
		log.Println("handling /yourproblem call")
		handlers.HandleYourProblem(w, r)
		return
	}

	if url == "/myproblem" {
		log.Println("handling /myproblem call")
		handlers.HandleMyProblem(w, r)
		return
	}

	if strings.HasPrefix(url, "/httpbin") {
		log.Println("handling /httpbin call")
		handlers.HandleHTTPBin(w, r)
		log.Println("handled /httpbin call")
		return
	}

	log.Println("handling other call")
	handlers.HandleOther(w, r)
}
