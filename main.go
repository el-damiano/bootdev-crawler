package main

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/el-damiano/bootdev-crawler/internal/goquery"
)

type Parser interface {
	GetFirstElement(html, element string) string
	GetFirstText(html, element string) string
}

func main() {
	fmt.Println("Hello, World!")
}

func getFirstH1FromHTML(html string) string {
	parser := goquery.Parser{}
	heading := parser.GetFirstText(html, "h1")
	return heading
}

func getFirstParagraphFromHTML(html string) string {
	// NOTE: some blogposts have mulitple h1
	// to avoid overthinking let's just get the first one
	parser := goquery.Parser{}
	main := parser.GetFirstElement(html, "main")
	if main != "" {
		return parser.GetFirstText(main, "p")
	} else {
		return parser.GetFirstText(html, "p")
	}
}

// func getURLsFromHTML(htmlBody string, baseURL *url.URL) ([]string, error) {
// 	return nil, nil // TODO
// }

// Normalizes a URL.
//
// The URL may be relative or absolute, but not empty.
// It can also be a sentence, as that's not being checked.
func URLnormalize(urlRaw string) (string, error) {
	if urlRaw == "" {
		return "", errors.New("Empty URL")
	}

	urlParsed, err := url.Parse(urlRaw)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	urlFullPath := strings.ToLower(urlParsed.Host + urlParsed.Path)
	urlNormalized := strings.TrimRight(urlFullPath, "/")

	return urlNormalized, nil
}
