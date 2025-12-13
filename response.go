package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func sendRequestToBackEnd(request Request, backend Backend, client net.Conn) (net.Conn, error) {
	request.HTTPVersion = "HTTP/1.1"
	request.header["Host"] = strings.Split(backend.Address, ":")[0]

	address := getBackendAddress(backend)

	var conn net.Conn
	var err error

	if backend.IsHTTPS {
		conf := &tls.Config{
			InsecureSkipVerify: true,
		}
		conn, err = tls.Dial("tcp", address, conf)
	} else {
		conn, err = net.Dial("tcp", address)
	}
	if err != nil {
		log.Printf("error : %v", err)
		return nil, err
	}
	requestString := fmt.Sprintf(
		"%s %s %s\r\n"+
			"Host: %s\r\n"+
			"User-Agent: %s\r\n"+
			"Accept: %s\r\n"+
			"Connection: keep-alive\r\n"+
			"\r\n",
		request.RequestType, request.Path, request.HTTPVersion,
		request.header["Host"],
		request.header["User-Agent"],
		request.header["Accept"],
	)

	_, err = conn.Write([]byte(requestString))
	if err != nil {
		return nil, err
	}
	return conn, err
}

func sendRequest(request Request, client net.Conn, backendIndex int, setCookie bool, backend Backend) {
	conn, err := sendRequestToBackEnd(request, backend, client)
	if err != nil {
		fmt.Println("backend error:", err)
		return
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)
	var stringBuilder strings.Builder

	ln, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	stringBuilder.WriteString(ln)
	header, err := getHeader(reader)
	if err != nil {
		fmt.Printf("error: %s", err)
		return
	}
	header["Connection"] = "close"
	headerToString(header, &stringBuilder)

	fmt.Fprintf(client, "%s", stringBuilder.String())

	_, err = io.Copy(client, reader)
	if err != nil {
		return
	}
	client.Close()
}
