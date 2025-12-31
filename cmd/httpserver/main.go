package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/uygardeniz/http-from-tcp/internal/headers"
	"github.com/uygardeniz/http-from-tcp/internal/request"
	"github.com/uygardeniz/http-from-tcp/internal/response"
	"github.com/uygardeniz/http-from-tcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, func(w *response.Writer, req *request.Request) *server.HandlerError {
		var statusCode response.StatusCode
		var body []byte

		if path, found := strings.CutPrefix(req.RequestLine.RequestTarget, "/httpbin"); found {
			targetURL := fmt.Sprintf("https://httpbin.org%s", path)

			resp, err := http.Get(targetURL)
			if err != nil {
				return &server.HandlerError{
					StatusCode: response.StatusInternalServerError,
					Message:    fmt.Sprintf("Failed to reach httpbin.org: %v", err),
				}
			}
			defer resp.Body.Close()

			w.WriteStatusLine(response.StatusOK)

			h := headers.NewHeaders()
			h.Set("Transfer-Encoding", "chunked")
			h.Set("Trailer", "X-Content-SHA256, X-Content-Length")
			h.Set("Connection", "close")
			h.Set("Content-Type", "application/json")
			w.WriteHeaders(h)

			var fullBody []byte

			buf := make([]byte, 1024)
			for {
				n, err := resp.Body.Read(buf)
				if n > 0 {

					fullBody = append(fullBody, buf[:n]...)
					_, writeErr := w.WriteChunkedBody(buf[:n])
					if writeErr != nil {
						log.Printf("Error writing chunk: %v", writeErr)
						break
					}
				}
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Printf("Error reading from httpbin.org: %v", err)
					break
				}
			}

			w.WriteChunkedBodyDone()

			hash := sha256.Sum256(fullBody)
			hashHex := hex.EncodeToString(hash[:])

			trailers := headers.NewHeaders()
			trailers.Set("X-Content-SHA256", hashHex)
			trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
			w.WriteTrailers(trailers)
			return nil
		} else if req.RequestLine.RequestTarget == "/yourproblem" {
			statusCode = response.StatusBadRequest
			body = []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
		} else if req.RequestLine.RequestTarget == "/myproblem" {
			statusCode = response.StatusInternalServerError
			body = []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
		} else {
			statusCode = response.StatusOK
			body = []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
		}

		w.WriteStatusLine(statusCode)
		headers := response.GetDefaultHeaders(len(body))
		headers.Override("Content-Type", "text/html")
		w.WriteHeaders(headers)
		w.WriteBody(body)
		return nil
	})

	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
