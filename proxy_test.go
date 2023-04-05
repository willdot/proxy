package proxy

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	source_addr = ":5000"
	dest_addr   = ":4000"
)

func TestDataForwarded(t *testing.T) {
	destConn, err := net.Listen("tcp", dest_addr)
	require.NoError(t, err)

	proxy, err := NewProxyServer(fmt.Sprintf("localhost%s", source_addr), fmt.Sprintf("localhost%s", dest_addr))
	require.NoError(t, err)

	t.Cleanup(func() {
		proxy.Close()
	})

	fin := make(chan struct{})
	go func() {
		conn, err := destConn.Accept()
		require.NoError(t, err)

		buf := make([]byte, 500)
		n, err := conn.Read(buf)

		require.NoError(t, err)
		assert.True(t, n > 0)
		assert.Equal(t, "hello world", string(buf[:n]))

		fin <- struct{}{}
	}()

	send, err := net.Dial("tcp", ":5000")
	require.NoError(t, err)

	_, err = send.Write([]byte("hello world"))
	require.NoError(t, err)

	<-fin
}
