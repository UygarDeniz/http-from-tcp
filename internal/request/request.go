package request

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/uygardeniz/http-from-tcp/internal/headers"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type parserState string

const (
	StateInitialized    parserState = "initialized"
	StateParsingHeaders parserState = "parsing_headers"
	StateParsingBody    parserState = "parsing_body"
	StateDone           parserState = "done"
)

const bufferSize = 8

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte

	state parserState
}

func NewRequest() *Request {
	return &Request{
		state:   StateInitialized,
		Headers: headers.NewHeaders(),
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := NewRequest()
	buf := make([]byte, bufferSize)
	readToIndex := 0

	for request.state != StateDone {

		if readToIndex == len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		n, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {

				if readToIndex > 0 {
					_, parseErr := request.parse(buf[:readToIndex])
					if parseErr != nil {
						return nil, parseErr
					}
				}

				if request.state == StateParsingBody {
					contentLengthStr := request.Headers.Get("Content-Length")
					if contentLengthStr != "" {
						contentLength, _ := strconv.Atoi(contentLengthStr)
						if len(request.Body) < contentLength {
							return nil, fmt.Errorf("body shorter than Content-Length")
						}
					}
				}
				request.state = StateDone
				break
			}
			return nil, err
		}
		readToIndex += n

		parsed, err := request.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		if parsed > 0 {
			copy(buf, buf[parsed:])
			readToIndex -= parsed
		}
	}

	return request, nil
}

func parseRequestLine(data string) (*RequestLine, int, error) {

	index := strings.Index(data, "\r\n")

	if index == -1 {
		return nil, 0, nil
	}

	allData := strings.Split(data, "\r\n")
	requestLine := allData[0]
	parts := strings.Split(requestLine, " ")

	if len(parts) != 3 {
		return nil, 0, fmt.Errorf("Wrong request Line")
	}

	method := parts[0]
	target := parts[1]
	httpVersion := strings.Split(parts[2], "/")[1]

	err := verifyMethod(method)
	if err != nil {
		return nil, 0, err
	}

	err = verifyHttpVersion(httpVersion)
	if err != nil {
		return nil, 0, err
	}

	return &RequestLine{httpVersion, target, method}, index + 2, nil

}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0

	for r.state != StateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		if n == 0 {
			break
		}
		totalBytesParsed += n
	}

	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case StateInitialized:
		requestLine, n, err := parseRequestLine(string(data))
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.state = StateParsingHeaders
		return n, nil

	case StateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.state = StateParsingBody
		}
		return n, nil

	case StateParsingBody:

		contentLengthStr := r.Headers.Get("Content-Length")
		if contentLengthStr == "" {
			r.state = StateDone
			return 0, nil
		}

		contentLength, err := strconv.Atoi(contentLengthStr)
		if err != nil {
			return 0, fmt.Errorf("invalid Content-Length: %s", contentLengthStr)
		}

		r.Body = append(r.Body, data...)

		if len(r.Body) > contentLength {
			return 0, fmt.Errorf("body length exceeds Content-Length")
		}

		if len(r.Body) == contentLength {
			r.state = StateDone
		}

		return len(data), nil

	case StateDone:
		return 0, fmt.Errorf("error: trying to read data in a done state")

	default:
		return 0, fmt.Errorf("error: unknown state")
	}
}

func verifyMethod(method string) error {
	allowedMetods := [5]string{"GET", "POST", "PUT", "PATCH", "DELETE"}

	for _, m := range allowedMetods {
		if m == method {
			return nil
		}
	}

	return fmt.Errorf("invalid method: %s", method)
}

func verifyHttpVersion(httpStr string) error {
	if httpStr != "1.1" {
		return fmt.Errorf("Unsupported http version")
	}

	return nil
}
