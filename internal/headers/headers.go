package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Get(key string) string {
	return h[strings.ToLower(key)]
}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)

	if v, ok := h[key]; ok {
		h[key] = fmt.Sprintf("%s,%s", v, value)
	} else {
		h[key] = value
	}
}

func (h Headers) Override(key, value string) {
	h[strings.ToLower(key)] = value
}

func (h Headers) Delete(key string) {
	key = strings.ToLower(key)
	delete(h, key)
}

func isValidTokenChar(c byte) bool {

	if c >= 'A' && c <= 'Z' {
		return true
	}

	if c >= 'a' && c <= 'z' {
		return true
	}

	if c >= '0' && c <= '9' {
		return true
	}

	switch c {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	}
	return false
}

func validateFieldName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("invalid header: empty field name")
	}
	for i := 0; i < len(name); i++ {
		if !isValidTokenChar(name[i]) {
			return fmt.Errorf("invalid header: invalid character in field name")
		}
	}
	return nil
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {

	crlfIndex := bytes.Index(data, []byte("\r\n"))
	if crlfIndex == -1 {
		return 0, false, nil
	}

	if crlfIndex == 0 {
		return 2, true, nil
	}

	headerLine := string(data[:crlfIndex])

	colonIndex := strings.Index(headerLine, ":")
	if colonIndex == -1 {
		return 0, false, fmt.Errorf("invalid header: no colon found")
	}

	fieldName := headerLine[:colonIndex]

	trimmedFieldName := strings.TrimSpace(fieldName)
	if strings.HasSuffix(fieldName, " ") || strings.HasSuffix(fieldName, "\t") {
		return 0, false, fmt.Errorf("invalid header: space before colon")
	}

	if fieldName != trimmedFieldName && len(fieldName) > 0 {
		lastChar := fieldName[len(fieldName)-1]
		if lastChar == ' ' || lastChar == '\t' {
			return 0, false, fmt.Errorf("invalid header: space before colon")
		}
	}

	fieldName = strings.TrimLeft(fieldName, " \t")

	if strings.TrimRight(fieldName, " \t") != fieldName {
		return 0, false, fmt.Errorf("invalid header: space before colon")
	}

	if err := validateFieldName(fieldName); err != nil {
		return 0, false, err
	}

	fieldValue := strings.TrimSpace(headerLine[colonIndex+1:])

	h.Set(fieldName, fieldValue)

	return crlfIndex + 2, false, nil
}
