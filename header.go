package main

import (
	"bufio"
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func formatToHTTPForm(key string) string {

	key = strings.ReplaceAll(key, " ", "-")
	key = strings.ToLower(key)
	caser := cases.Title(language.English)

	return caser.String(key)
}

func headerToString(header map[string]string, stringBuilder *strings.Builder) {

	for v, n := range header {
		stringBuilder.Write([]byte(v))
		stringBuilder.Write([]byte(": "))
		stringBuilder.Write([]byte(n))
		stringBuilder.Write([]byte("\r\n"))
	}
	stringBuilder.Write([]byte("\r\n"))
}

func getHeader(reader *bufio.Reader) (map[string]string, error) {
	header := make(map[string]string)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.Trim(line, "\r\n")
		if line == "" {
			return header, nil
		}
		lineSplit := strings.SplitN(line, ":", 2)
		if len(lineSplit) != 2 {
			return nil, fmt.Errorf("invalid header format")
		}
		rawkey := strings.TrimSpace(lineSplit[0])
		if strings.Contains(rawkey, " ") {
			return nil, fmt.Errorf("invalid header format")
		}
		key := formatToHTTPForm(rawkey)
		data := strings.TrimSpace(lineSplit[1])

		header[key] = data
	}
}
