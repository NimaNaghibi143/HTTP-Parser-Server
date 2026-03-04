package request

// a simple test just to see everything is set up!

import (
	"strings"
	"testing"
)

func TestRequestLineParse(t *testing.T) {
	// Test: a correct GET request line
	r, err := RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\nHost: localhost:42069"))
	require.NoError(t, err)
	require.NoNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: correct GET request line with path
	r, err = RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\nHost: localhost:42069"))
	require.NoError(t, err)
	require.NoNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	//Test: invalid number of parts in request line
	_, err = RequestFromReader(strings.NewReader("/coffee HTTP/1.1\r\nHost: localhost:42069"))
	require.Error(t, err)
}
