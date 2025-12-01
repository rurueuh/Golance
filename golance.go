package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net"
	"strconv"
	"strings"
	"sync"
)

type Header struct {
	RequestType string
	Path        string
	HTTPVersion string

	UserAgent  string
	Host       string
	Accept     string
	Connection string
}

type Backend struct {
	Address string
	IsHTTPS bool
}

var backends = []Backend{
	{Address: "example.com:443", IsHTTPS: true},
	{Address: "rurueuh.fr:443", IsHTTPS: true},
}

func getBackendAddress(b Backend) string {
	if !strings.Contains(b.Address, ":") {
		port := 80
		if b.IsHTTPS {
			port = 443
		}
		return fmt.Sprintf("%s:%s", b.Address, port)
	}
	return b.Address
}

func sendRequest(header Header, client net.Conn, backendIndex int, setCookie bool, backend Backend) {
	defer client.Close()

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

	responseHeader := headerBuffer.String()
	if setCookie && isHeaderEnd {
		fle := strings.Index(responseHeader, "\r\n")
		if fle != -1 {
			cookieHeader := fmt.Sprintf("Set-Cookie: LB_NODE=%d; PATH=/; Max-Age=60\r\n", backendIndex)
			newResponse := responseHeader[:fle+2] + cookieHeader + responseHeader[fle+2:]
			client.Write([]byte(newResponse))
		} else {
			client.Write([]byte(responseHeader))
		}
	} else {
		client.Write([]byte(responseHeader))
	}
	_, err = io.Copy(client, reader)
	if err != nil {
		return
	}
}

func headerParser(reader *bufio.Reader) (Header, int, bool, error) {
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return Header{}, 0, false, err
	}
	words := strings.Fields(requestLine)
	if len(words) < 3 {
		return Header{}, 0, false, fmt.Errorf("bad http format")
	}

	var selectedIndex int
	foundCookie := false

	var currentHeader strings.Builder
	currentHeader.WriteString(requestLine)

	for {
		line, err := reader.ReadString('\n')
		if err != nil || line == "\r\n" || line == "\n" {
			currentHeader.WriteString(line)
			break
		}
		currentHeader.WriteString(line)
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

	var header Header
	header.RequestType = words[0]
	header.Path = words[1]
	header.HTTPVersion = words[2]

	header.UserAgent = "GoLanceProxy"
	header.Accept = "*/*"
	header.Connection = "close"
	header.Host = strings.Split(backends[selectedIndex].Address, ":")[0]

	return header, selectedIndex, setCookie, nil
}

func handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)

	header, selectedIndex, setCookie, err := headerParser(reader)
	if err != nil {
		conn.Close()
		return
	}

	backend := backends[selectedIndex]
	go sendRequest(header, conn, selectedIndex, setCookie, backend)
}

func listenHTTP(port string) {
	ln, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("error on open http %s: %v", port, err)
		return
	}
	defer ln.Close()

	fmt.Println("HTTP server running on port", port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn)
	}
}

func listenHTTPS(port string) {
	server, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		log.Fatalf("cert.pem or key.pem is invalid : %v", err)
	}

	config := &tls.Config{Certificates: []tls.Certificate{server}}
	ln, err := tls.Listen("tcp", port, config)
	if err != nil {
		log.Fatalf("error on open https %s: %v", port, err)
	}
	defer ln.Close()

	fmt.Println("HTTPS server running on port", port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn)
	}
}

func main() {

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		listenHTTP(":8080")
	}()

	go func() {
		defer wg.Done()
		listenHTTPS(":8443")
	}()

	wg.Wait()
}
