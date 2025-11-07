# HTTP from TCP

## Overview

This project is designed to take a TCP connection and parse it into a valid HTTP request. Written in Go, this library aims to help gain a deeper understanding of the inner workings of HTTP requests and the process of parsing TCP connections.

## Features

- **TCP Connection Handling**: Accepts raw TCP connections and reads data.
- **HTTP Request Parsing**: Parses the raw data into structured HTTP request objects.
- **Error Handling**: Provides robust error handling for malformed requests.

## Reasoning
I decided to write this project as a way to gain a deeper understanding of the inner workings of HTTP requests and the process of parsing TCP connections. In daily working and development, we are lucky enough to have access to a variety of tools and libraries that can help us with this task. However, I wanted to explore the process of parsing TCP connections and HTTP requests in a more hands-on manner. By completing this project, I hope to gain a better understanding of the underlying mechanics of HTTP requests and the process of parsing TCP connections, and the flow of data from client to server, and back to client again in the form of a response.

### Prerequisites
- Go 1.16 or later
- Basic understanding of TCP and HTTP protocols

