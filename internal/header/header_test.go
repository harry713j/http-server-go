package header

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test: Valid single header
func TestValidSingleHeader(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)
}

// Test: Invalid spacing header
func TestInvalidSpacingHeader(t *testing.T) {
	headers := NewHeaders()
	data := []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

// Test: "Valid single header with extra whitespace"
func TestValidSingleHeaderWithExtraSpace(t *testing.T) {
	headers := NewHeaders()
	data := []byte("      Host: localhost:42069       \r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 36, n)
	assert.False(t, done)
}

// Test; "Valid 2 headers with existing headers" and Valid done
func TestValidTwoHeaders(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// 2nd header
	data = []byte("Accept-Language: en-US\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "en-US", headers["accept-language"])
	assert.Equal(t, 24, n)
	assert.False(t, done)

	data = []byte("\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.True(t, done)
}

// Test: Valid single header (mixed case)
func TestValidSingleHeaderMixedCase(t *testing.T) {
	headers := NewHeaders()
	data := []byte("HoSt: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"]) // key must be lowercase
	assert.Equal(t, 23, n)
	assert.False(t, done)
}

// Test: Invalid character in header key
func TestInvalidCharacterInHeaderKey(t *testing.T) {
	headers := NewHeaders()
	data := []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

// Test: "Duplicate header should append value"
func TestDuplicateHeaderAppendsValue(t *testing.T) {
	headers := NewHeaders()

	// First header
	n, done, err := headers.Parse([]byte("Accept: text/html\r\n"))
	require.NoError(t, err)
	assert.Equal(t, 19, n)
	assert.False(t, done)

	// Second header
	n, done, err = headers.Parse([]byte("Accept: application/json\r\n\r\n"))
	require.NoError(t, err)
	assert.Equal(t, "text/html, application/json", headers["accept"])
	assert.Equal(t, 26, n)
	assert.False(t, done)
}
