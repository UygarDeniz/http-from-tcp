package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHeaders_ValidSingleHeader(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	// Use Get for case-insensitive access
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, 23, n)
	assert.False(t, done)
}

func TestParseHeaders_InvalidSpacingHeader(t *testing.T) {
	headers := NewHeaders()
	data := []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

func TestParseHeaders_CaseInsensitiveKeys(t *testing.T) {
	// Test that mixed case keys are stored as lowercase
	headers := NewHeaders()
	data := []byte("Content-Type: application/json\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 32, n)
	assert.False(t, done)
	// Should be accessible via Get (case-insensitive)
	assert.Equal(t, "application/json", headers.Get("Content-Type"))

	// Parse another header with different casing
	headers2 := NewHeaders()
	data2 := []byte("CONTENT-LENGTH: 42\r\n")
	n2, done2, err2 := headers2.Parse(data2)
	require.NoError(t, err2)
	assert.Equal(t, 20, n2)
	assert.False(t, done2)
	// Should be accessible via Get (case-insensitive)
	assert.Equal(t, "42", headers2.Get("Content-Length"))
}

func TestParseHeaders_InvalidCharacterInKey(t *testing.T) {
	headers := NewHeaders()
	// © is not a valid token character
	data := []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

func TestParseHeaders_InvalidCharacterSpace(t *testing.T) {
	headers := NewHeaders()
	// Space in header name is invalid
	data := []byte("Ho st: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

func TestParseHeaders_ValidSpecialChars(t *testing.T) {
	// Test valid special characters in field name
	headers := NewHeaders()
	data := []byte("X-Custom-Header_123: value\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 28, n)
	assert.False(t, done)
	assert.Equal(t, "value", headers.Get("X-Custom-Header_123"))
}

func TestParseHeaders_EndOfHeaders(t *testing.T) {
	headers := NewHeaders()
	data := []byte("\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.True(t, done)
}

func TestParseHeaders_NoCRLF(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Host: localhost:42069")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

func TestParseHeaders_EmptyFieldName(t *testing.T) {
	headers := NewHeaders()
	data := []byte(": value\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}
