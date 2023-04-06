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

	w io.Writer

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

func (p *ProxyServer) AddAdditionalWriter(w io.Writer) {
	p.w = w
}

// Close will close down any started connections
func (p *ProxyServer) Close() error {
	return p.listener.Close()
}

func (p *ProxyServer) run() error {
	go func() {
		for {
			conn, err := p.listener.Accept()
			if err != nil {
				// TODO: work out if there's a way to stop waiting for accept if we are closing the conns down.
				fmt.Printf("error accepting source conn: %v\n", err)
				return
			}

			forwardConn := forwardConnection{
				sourceConn:      conn,
				destinationAddr: p.destAddr,
				w:               p.w,
			}

			go forwardConn.process()
		}
	}()

	return nil
}

type forwardConnection struct {
	sourceConn      net.Conn
	destinationAddr string
	w               io.Writer
}

func (f *forwardConnection) process() {
	defer f.sourceConn.Close()

	destConn, err := net.Dial("tcp", f.destinationAddr)
	if err != nil {
		fmt.Sprintf("failed to dial destination address: %v\n", err)
	}

	writer := f.getWriter(destConn)

	for {
		_, err = io.CopyN(writer, f.sourceConn, 512)
		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Printf("error copying from source to destination: %v\n", err)
			break
		}
	}

}

func (f *forwardConnection) getWriter(destConn net.Conn) io.Writer {
	if f.w == nil {
		return destConn
	}

	return io.MultiWriter(destConn, f.w)
}
