package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type header_t struct {
	RequestType string
	Path        string
	HTTPVersion string

	UserAgent  string
	Host       string
	Accept     string
	Connection string
}

func sendRequest(header header_t) {
	ipToSend := fmt.Sprintf("%s:80", header.Host)
	conn, err := net.Dial("tcp", ipToSend)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer conn.Close()
	requestString := fmt.Sprintf(
		"%s %s %s\r\n"+
			"Host: %s\r\n"+
			"User-Agent: %s\r\n"+
			"Accept: %s\r\n"+
			"Connection: %s\r\n"+
			"\r\n",

		header.RequestType, header.Path, header.HTTPVersion,
		header.Host,
		header.UserAgent,
		header.Accept,
		header.Connection,
	)

	fmt.Printf("data send: %s", requestString)

	_, err = conn.Write([]byte(requestString))
	if err != nil {
		fmt.Println(err)
		return
	}

	buffer := make([]byte, 10240)
	_, err = conn.Read(buffer)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("data: %s", buffer)
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	words := strings.Fields(requestLine)

	if len(words) < 3 {
		return
	}

	var header header_t
	header.RequestType = words[0]
	header.Path = words[1]
	header.HTTPVersion = words[2]

	header.UserAgent = "GoLanceProxy"
	header.Accept = "*/*"
	header.Connection = "keep-alive"
	header.Host = "example.com"

	sendRequest(header)

	fmt.Println(header)
}

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ln.Close()
	fmt.Println("listen on port 8080")

	for {
		conn, err := ln.Accept()
		if err != nil {
			println(err)
			continue
		}
		go handleConnection(conn)
	}
}
