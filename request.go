package main

import (
	"bufio"
	"fmt"
	"math/rand/v2"
	"net"
	"slices"
	"strconv"
	"strings"
)

func handleConnection(conn net.Conn) error {
	request, err := CreateRequest(conn)
	if err != nil {
		fmt.Println(err)
		return err
	}

	selectedIndex := -1
	foundCookie := false
	cookieHeader := request.header["Cookie"]
	cookies := strings.SplitSeq(cookieHeader, ";")
	for v := range cookies {
		if v == "" {
			continue
		}
		cookie := strings.Split(v, "=")
		if len(cookie) != 2 {
			return fmt.Errorf("bad request")
		}
		key := strings.TrimSpace(cookie[0])
		if key == "LB_GOLANCE" {
			selectedIndex, _ = strconv.Atoi(cookie[1])
			foundCookie = true
			break
		}
	}

	if !foundCookie || len(backends) < selectedIndex || len(backends) > selectedIndex {
		selectedIndex = rand.IntN(len(backends))
	}
	backend := backends[selectedIndex]
	sendRequest(request, conn, selectedIndex, foundCookie, backend)
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
