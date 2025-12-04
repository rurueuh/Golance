package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"slices"
	"strings"
	"sync"
)

type Backend struct {
	Address string
	IsHTTPS bool
}

var backends = []Backend{
	{Address: "example.com:443", IsHTTPS: true},
	{Address: "example.com:80", IsHTTPS: false},
	{Address: "rurueuh.fr:443", IsHTTPS: true},
}

var validateMethod = []string{"OPTIONS", "GET", "HEAD", "POST", "PUT", "DELETE", "TRACE", "CONNECT"}

func getBackendAddress(b Backend) string {
	if !strings.Contains(b.Address, ":") {
		port := 80
		if b.IsHTTPS {
			port = 443
		}
		return fmt.Sprintf("%s:%d", b.Address, port)
	}
	return b.Address
}

type Request struct {
	RequestType string
	Path        string
	HTTPVersion string

	header map[string]string

	body string
}

func sendRequest(request Request, client net.Conn, backendIndex int, setCookie bool, backend Backend) {
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
		request.RequestType, request.Path, request.HTTPVersion,
		request.header["Host"],
		request.header["User-Agent"],
		request.header["Accept"],
		request.header["Connection"],
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

func handleConnection(conn net.Conn) error {
	request, err := CreateRequest(conn)
	if err != nil {
		fmt.Println(err)
		return err
	}

	lbValue := ""
	cookieHeader := request.header["Cookie"]
	cookieHeader = strings.TrimSpace(cookieHeader)
	cookies := strings.SplitSeq(cookieHeader, ";")
	for v := range cookies {
		if v == "" {
			continue
		}
		cookie := strings.Split(v, ":")
		if len(cookie) != 2 {
			return fmt.Errorf("bad request")
		}
		key := cookie[0]
		fmt.Printf("key %s", key)
		if key == "LB_GOLANCE" {
			lbValue = cookie[1]
		}
	}

	// backend := backends[selectedIndex]
	// go sendRequest(header, conn, selectedIndex, setCookie, backend)
	return nil
}

func CreateRequest(conn net.Conn) (Request, error) {
	var (
		request Request
		err     error
		reader  *bufio.Reader
	)

	reader = bufio.NewReader(conn)

	firstLine, err := reader.ReadString('\n')
	if err != nil {
		return Request{}, err
	}
	firstLineSplit := strings.Split(firstLine, " ")
	if len(firstLineSplit) != 3 {
		return Request{}, fmt.Errorf("bad request")
	}

	if !slices.Contains(validateMethod, firstLineSplit[0]) {
		return Request{}, fmt.Errorf("bad request")
	}

	request.RequestType = firstLineSplit[0]
	request.Path = firstLineSplit[1]
	request.HTTPVersion = firstLineSplit[2]

	request.header, err = getHeader(reader)
	if err != nil {
		conn.Close()
		return Request{}, err
	}
	return request, nil
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
