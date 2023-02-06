package main

import (
	"errors"
	"strings"
)

func parseHTML(s string) ([]string, []string, error) {
	images := make([]string, 0, 1)
	links := make([]string, 0, 2)

	for len(s) > 0 {
		// Find HTML tag start
		pos := strings.IndexByte(s, '<')
		if pos == -1 {
			break
		}
		s = s[pos+1:]

		s = skipSpaces(s)

		// Find end of token
		pos = strings.IndexAny(s, " \t\n\r>")
		if pos == -1 {
			return nil, nil, errors.New("invalid html")
		}

		token := s[:pos]
		s = s[pos:]

		// Skip comments
		if token == "!--" {
			pos := strings.Index(s, "-->")
			if pos == -1 {
				break
			}
			s = s[pos+3:]
			continue
		}

		// Find end of HTML tag
		pos = strings.IndexByte(s, '>')
		if pos == -1 {
			return nil, nil, errors.New("invalid html")
		}
		attrs := s[:pos]
		s = s[pos+1:]

		if attrs != "" {
			switch toLower(token) {
			case "a":
				pos := strings.Index(attrs, "href")
				if pos != -1 {
					attrs = skipSpaces(attrs[pos+4:])
					if attrs[0] == '=' {
						attrs = skipSpaces(attrs[1:])
					}
					if attrs[0] != '"' {
						return nil, nil, errors.New("invalid html")
					}

					pos = strings.IndexByte(attrs[1:], '"')
					if pos == -1 {
						return nil, nil, errors.New("invalid html")
					}
					if pos > 0 {
						links = append(links, attrs[1:pos+1])
					}
				}
			case "img":
				pos := strings.Index(attrs, "src")
				if pos != -1 {
					attrs = skipSpaces(attrs[pos+3:])
					if attrs[0] == '=' {
						attrs = skipSpaces(attrs[1:])
					}
					if attrs[0] != '"' {
						return nil, nil, errors.New("invalid html")
					}

					pos = strings.IndexByte(attrs[1:], '"')
					if pos == -1 {
						return nil, nil, errors.New("invalid html")
					}
					if pos > 0 {
						images = append(images, attrs[1:pos+1])
					}
				}
			}
		}
	}

	return links, images, nil
}

func skipSpaces(s string) string {
	for len(s) > 0 {
		switch s[0] {
		case ' ', '\t', '\r', '\n':
			s = s[1:]
		default:
			return s
		}
	}

	return s
}

func toLower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c - 'A' + 'a'
		}
	}

	return string(b)
}
