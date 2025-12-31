package response

import (
	"fmt"
	"io"

	"github.com/uygardeniz/http-from-tcp/internal/headers"
)

type StatusCode int

type Writer struct {
	writer io.Writer
}

func NewWriter(writer io.Writer) *Writer {
	return &Writer{
		writer: writer,
	}
}

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	statusLine := []byte{}
	switch statusCode {
	case StatusOK:
		statusLine = []byte("HTTP/1.1 200 OK\r\n")
	case StatusBadRequest:
		statusLine = []byte("HTTP/1.1 400 Bad Request\r\n")
	case StatusInternalServerError:
		statusLine = []byte("HTTP/1.1 500 Internal Server Error\r\n")
	default:
		statusLine = []byte(fmt.Sprintf("HTTP/1.1 %d \r\n", statusCode))
	}

	_, err := w.Write(statusLine)
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, val := range headers {
		_, err := fmt.Fprintf(w, "%s: %s\r\n", key, val)
		if err != nil {
			return err
		}
	}

	_, err := w.Write([]byte("\r\n"))
	return err
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	statusLine := []byte{}
	switch statusCode {
	case StatusOK:
		statusLine = []byte("HTTP/1.1 200 OK\r\n")
	case StatusBadRequest:
		statusLine = []byte("HTTP/1.1 400 Bad Request\r\n")
	case StatusInternalServerError:
		statusLine = []byte("HTTP/1.1 500 Internal Server Error\r\n")
	default:
		statusLine = []byte(fmt.Sprintf("HTTP/1.1 %d \r\n", statusCode))
	}

	_, err := w.writer.Write(statusLine)
	return err
}
func (w *Writer) WriteHeaders(headers headers.Headers) error {
	for key, val := range headers {
		_, err := fmt.Fprintf(w.writer, "%s: %s\r\n", key, val)
		if err != nil {
			return err
		}
	}

	_, err := w.writer.Write([]byte("\r\n"))
	return err
}
func (w *Writer) WriteBody(p []byte) (int, error) {
	n, err := w.writer.Write(p)

	return n, err
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {

	chunkSize := fmt.Sprintf("%x\r\n", len(p))
	_, err := w.writer.Write([]byte(chunkSize))
	if err != nil {
		return 0, err
	}

	n, err := w.writer.Write(p)
	if err != nil {
		return n, err
	}

	_, err = w.writer.Write([]byte("\r\n"))
	if err != nil {
		return n, err
	}

	return n, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	return w.writer.Write([]byte("0\r\n"))
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	for key, val := range h {
		_, err := fmt.Fprintf(w.writer, "%s: %s\r\n", key, val)
		if err != nil {
			return err
		}
	}
	_, err := w.writer.Write([]byte("\r\n"))
	return err
}
