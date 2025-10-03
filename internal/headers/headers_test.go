package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestInvalidSpacingHeader(t *testing.T) {
	headers := NewHeaders()
	data := []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

func TestValidSingleHeaderWithSpacing(t *testing.T) {
	headers := NewHeaders()
	data := []byte("       Host: localhost:42069                           \r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 57, n)
	assert.False(t, done)
}

func TestValid2HeadersWithExisting(t *testing.T) {
	var headers Headers = map[string]string{"host": "localhost:42069"}
	data := []byte("User-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, "curl/7.81.0", headers["user-agent"])
	assert.Equal(t, 25, n)
	assert.False(t, done)
}

func TestValidDone(t *testing.T) {
	headers := NewHeaders()
	data := []byte("\r\n a bunch of other stuff")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Empty(t, headers)
	assert.Equal(t, 2, n)
	assert.True(t, done)
}

func TestInvalidHeaderKey(t *testing.T) {
	headers := NewHeaders()
	data := []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

func TestValidSingleHeaderKeyWithMultipleValues(t *testing.T) {
	var headers Headers = map[string]string{"host": "What even value"}
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "What even value, localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)
}
