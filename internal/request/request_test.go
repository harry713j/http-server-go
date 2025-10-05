package request

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

// Test: Good GET Request line
func TestGoodGetRequestline(t *testing.T) {
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}

	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
}

// Test: Good GET Request line with path
func TestGoodGetRequestlineWithPath(t *testing.T) {
	reader := &chunkReader{
		data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 1,
	}

	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
}

// Test: Invalid number of parts in request line
func TestInvalidRequestline(t *testing.T) {
	reader := &chunkReader{
		data:            "/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 2,
	}

	_, err := RequestFromReader(reader)
	require.Error(t, err)
}

// Test: Good POST request
func TestGoodPostRequestWithPath(t *testing.T) {
	reader := &chunkReader{
		data:            "POST /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\nContent-Length: 21\r\n\r\n{\"flavour\":\"darkmode\"}",
		numBytesPerRead: 1,
	}

	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "POST", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
}

// Test: Invalid method
func TestInvalidMethod(t *testing.T) {
	reader := &chunkReader{
		data:            "Get /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}

	r, err := RequestFromReader(reader)
	require.Error(t, err)
	require.Nil(t, r)
}

// Test: Invalid version
func TestInvalidVersion(t *testing.T) {
	reader := &chunkReader{
		data:            "GET /coffee HTTP/2.2\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}

	r, err := RequestFromReader(reader)
	require.Error(t, err)
	require.Nil(t, r)
}

func TestRequestParseHeaders(t *testing.T) {
	t.Run("Standard Headers", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 3,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "localhost:42069", r.Headers["host"])
		assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
		assert.Equal(t, "*/*", r.Headers["accept"])
	})

	t.Run("Malformed Header", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
			numBytesPerRead: 3,
		}
		r, err := RequestFromReader(reader)
		require.Error(t, err)
		assert.Nil(t, r)
	})

	t.Run("Empty Headers", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\n\r\n",
			numBytesPerRead: 2,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Empty(t, r.Headers)
	})

	t.Run("Duplicate Headers", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nAccept: text/html\r\nAccept: application/json\r\n\r\n",
			numBytesPerRead: 5,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		assert.Equal(t, "text/html, application/json", r.Headers["accept"])
	})

	t.Run("Case Insensitive Headers", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHOST: localhost\r\nhost: example.com\r\n\r\n",
			numBytesPerRead: 6,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		assert.Equal(t, "localhost, example.com", r.Headers["host"])
	})

	t.Run("Missing End of Headers", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost: localhost",
			numBytesPerRead: 5,
		}
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	})
}

func TestStandardBody(t *testing.T) {
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 13\r\n" +
			"\r\n" +
			"hello world!\n",
		numBytesPerRead: 3,
	}

	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "hello world!\n", string(r.Body))
}

func TestBodyShorterThanReportedLength(t *testing.T) {
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 20\r\n" +
			"\r\n" +
			"partial content",
		numBytesPerRead: 3,
	}

	r, err := RequestFromReader(reader)
	require.Error(t, err)
	assert.Nil(t, r)
}

func TestEmptyBodyWithZeroLength(t *testing.T) {
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 0\r\n" +
			"\r\n",
		numBytesPerRead: 3,
	}

	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	assert.Empty(t, r.Body)
}
