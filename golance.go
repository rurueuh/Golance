package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type header_t struct {
	requestType string
	path        string
	httpVersion string
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
	header.requestType = words[0]
	header.path = words[1]
	header.httpVersion = words[2]

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
