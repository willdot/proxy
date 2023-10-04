package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/willdot/proxy"
)

const (
	destAddr   = ":5000"
	sourceAddr = ":9000"
)

func main() {
	go createDestServer()

	proxy, err := proxy.NewProxyServer(sourceAddr, destAddr)
	if err != nil {
		panic(err)
	}
	defer proxy.Close()

	go sendData("hello")
	go sendData("goodbye")

	time.Sleep(time.Second)
}

func createDestServer() {
	lis, err := net.Listen("tcp", "localhost"+destAddr)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := lis.Accept()
		if err != nil {
			panic(err)
		}

		go handleDestConn(conn)
	}
}
func handleDestConn(conn net.Conn) {
	defer func() {
		conn.Close()
	}()

	b := new(bytes.Buffer)

	for {
		_, err := io.CopyN(b, conn, 1)
		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Printf("failed to read from conn: %s\n", err)
			return
		}
	}
	fmt.Printf("msg recevived: %s\n", b.Bytes())
}

func sendData(msg string) {
	conn, err := net.Dial("tcp", "localhost"+sourceAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("sending " + msg)
	_, err = conn.Write([]byte(msg))
	if err != nil {
		panic(err)
	}
}
