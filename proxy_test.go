package proxy_test

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/willdot/proxy"
)

const (
	source_addr = ":5000"
	dest_addr   = ":4000"
)

func TestDataForwarded(t *testing.T) {
	// create a destination conn where we want our data to be
	// forwarded to and start listening for connections / data
	destConn, err := net.Listen("tcp", dest_addr)
	require.NoError(t, err)

	t.Cleanup(func() {
		destConn.Close()
	})

	fin := make(chan struct{})
	go func() {
		conn, err := destConn.Accept()
		require.NoError(t, err)

		buf := make([]byte, 500)
		n, err := conn.Read(buf)

		require.NoError(t, err)
		assert.Equal(t, "hello world", string(buf[:n]))

		fin <- struct{}{}
	}()

	// create a new proxy server
	p, err := proxy.NewProxyServer("localhost"+source_addr, "localhost"+dest_addr)
	require.NoError(t, err)

	t.Cleanup(func() {
		p.Close()
	})

	// send some data to the source address of the proxy server which should then be received
	// by the destination conn created at the start
	send, err := net.Dial("tcp", source_addr)
	require.NoError(t, err)

	_, err = send.Write([]byte("hello world"))
	require.NoError(t, err)

	// wait for the destination connection to receive and assert data
	select {
	case <-fin:
		return
	case <-time.After(time.Second * 10):
		t.Fatal("test timed out waiting")
	}
}

func TestDataWrittenToAdditionalWriterAsWellAsDestination(t *testing.T) {
	// create a destination conn where we want our data to be
	// forwarded to and start listening for connections / data
	destConn, err := net.Listen("tcp", dest_addr)
	require.NoError(t, err)

	t.Cleanup(func() {
		destConn.Close()
	})

	fin := make(chan struct{})
	go func() {
		conn, err := destConn.Accept()
		require.NoError(t, err)

		buf := make([]byte, 500)
		n, err := conn.Read(buf)

		require.NoError(t, err)
		assert.Equal(t, "hello world", string(buf[:n]))

		fin <- struct{}{}
	}()

	// create a new proxy server
	p, err := proxy.NewProxyServer("localhost"+source_addr, "localhost"+dest_addr)
	require.NoError(t, err)

	t.Cleanup(func() {
		p.Close()
	})

	buf := bytes.NewBuffer(nil)

	p.AddAdditionalWriter(buf)

	// send some data to the source address of the proxy server which should then be received
	// by the destination conn created at the start
	send, err := net.Dial("tcp", source_addr)
	require.NoError(t, err)

	_, err = send.Write([]byte("hello world"))
	require.NoError(t, err)

	// wait for the destination connection to receive and assert data
	select {
	case <-fin:
		break
	case <-time.After(time.Second * 10):
		t.Fatal("test timed out waiting")
	}

	// check the data was also written to the buffer
	assert.Equal(t, "hello world", string(buf.Bytes()))
}
