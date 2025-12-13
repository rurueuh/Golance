package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

type Backend struct {
	Address string
	IsHTTPS bool
}

var backends = []Backend{
	{Address: "rurueuh.fr:443", IsHTTPS: true},
	{Address: "example.com:443", IsHTTPS: true},
	{Address: "example.com:80", IsHTTPS: false},
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
