package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"strconv"
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

func sendRequest(header header_t, client net.Conn, backendIndex int, setCookie bool) {
	defer client.Close()
	ipToSend := fmt.Sprintf("%s:80", header.Host)
	conn, err := net.Dial("tcp", ipToSend)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer conn.Close()
	fmt.Printf("request : %s goto %s\n", header.Path, header.Host)
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

	_, err = conn.Write([]byte(requestString))
	if err != nil {
		fmt.Println(err)
		return
	}

	reader := bufio.NewReader(conn)

	var headerBuffer strings.Builder
	var isHeaderEnd bool
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		if line == "\r\n" || line == "\n" {
			isHeaderEnd = true
			headerBuffer.WriteString(line)
			break
		}
		headerBuffer.WriteString(line)
	}

	reponseHeader := headerBuffer.String()
	if setCookie && isHeaderEnd {
		fle := strings.Index(reponseHeader, "\r\n")
		if fle != -1 {
			cookieHeader := fmt.Sprintf("Set-Cookie: LB_NODE=%d; PATH=/; Max-Age=60\r\n", backendIndex)
			newResponse := reponseHeader[:fle+2] + cookieHeader + reponseHeader[fle+2:]
			client.Write([]byte(newResponse))
		}
	} else {
		client.Write([]byte(reponseHeader))
	}
	_, err = io.Copy(client, reader)
	if err != nil {
		return
	}
}

func handleConnection(conn net.Conn, backends []string) {
	reader := bufio.NewReader(conn)

	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	words := strings.Fields(requestLine)

	if len(words) < 3 {
		return
	}

	var selectedIndex int
	foundCookie := false

	for {
		line, err := reader.ReadString('\n')
		if err != nil || line == "\r\n" || line == "\n" {
			break
		}
		if strings.HasPrefix(line, "Cookie:") {
			if strings.Contains(line, "LB_NODE=") {
				parts := strings.Split(line, "LB_NODE=")
				if len(parts) > 1 {
					valStr := strings.Split(parts[1], ";")[0]
					valStr = strings.TrimSpace(valStr)

					idx, err := strconv.Atoi(valStr)
					if err == nil && idx >= 0 && idx < len(backends) {
						selectedIndex = idx
						foundCookie = true
					}
				}
			}
		}
	}

	setCookie := false
	if !foundCookie {
		selectedIndex = rand.IntN(len(backends))
		setCookie = true
	}

	var header header_t
	header.RequestType = words[0]
	header.Path = words[1]
	header.HTTPVersion = words[2]

	header.UserAgent = "GoLanceProxy"
	header.Accept = "*/*"
	header.Connection = "close"
	header.Host = backends[selectedIndex]

	go sendRequest(header, conn, selectedIndex, setCookie)
}

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ln.Close()
	fmt.Println("listen on port 8080")

	backends := []string{
		"example.com",
		"httpforever.com",
		"192.168.1.254",
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			println(err)
			continue
		}

		go handleConnection(conn, backends)
	}
}
