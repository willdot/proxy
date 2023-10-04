package proxy

import (
	"fmt"
	"io"
	"net"

	"github.com/pkg/errors"
)

// ProxyServer will listen on a source address and forward the data to the destination address
type ProxyServer struct {
	sourceAddr string
	destAddr   string

	listener net.Listener
}

// NewProxyServer will create a new proxy server and start listening for data on the source address
func NewProxyServer(sourceAddr, destAddr string) (*ProxyServer, error) {
	proxy := ProxyServer{
		sourceAddr: sourceAddr,
		destAddr:   destAddr,
	}

	sourceLis, err := net.Listen("tcp", sourceAddr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to listen on source address")
	}
	proxy.listener = sourceLis

	go proxy.run()

	return &proxy, nil
}

// Close will close down any started connections
func (p *ProxyServer) Close() error {
	return p.listener.Close()
}

func (p *ProxyServer) run() {

	for {
		conn, err := p.listener.Accept()
		if err != nil {
			fmt.Printf("error accepting source conn: %v\n", err)
			return
		}

		forwardConn := forwardConnection{
			sourceConn:      conn,
			destinationAddr: p.destAddr,
		}

		go forwardConn.process()
	}
}

type forwardConnection struct {
	sourceConn      net.Conn
	destinationAddr string
}

func (f *forwardConnection) process() {
	defer f.sourceConn.Close()

	destConn, err := net.Dial("tcp", f.destinationAddr)
	if err != nil {
		fmt.Sprintf("failed to dial destination address: %v\n", err)
	}

	b := make([]byte, 512)
	for {
		n, readErr := f.sourceConn.Read(b)
		if readErr != nil && readErr != io.EOF {
			fmt.Printf("failed to read from source conn: %v\n", readErr)
			break
		}

		_, writeErr := destConn.Write(b[:n])
		if writeErr != nil {
			fmt.Printf("failed to send data to destination: %v", writeErr)
			break
		}

		if readErr == io.EOF {
			break
		}
	}
}
