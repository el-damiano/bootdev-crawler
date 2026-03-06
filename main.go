package main

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/el-damiano/bootdev-crawler/internal/goquery"
)

func main() {
	fmt.Println("Hello, World!")
}

type PageData struct {
	URL            string
	Heading        string
	FirstParagraph string
	OutgoingLinks  []string
	ImageUrls      []string
}

func extractPageData(html, pageURL string) PageData {
	pageData := PageData{
		URL:            pageURL,
		Heading:        getFirstH1FromHTML(html),
		FirstParagraph: getFirstParagraphFromHTML(html),
		OutgoingLinks:  nil,
		ImageUrls:      nil,
	}

	baseURL, err := url.Parse(pageURL)
	if err != nil {
		return pageData
	}

	outgoingLinks, err := getUrlsFromHTML(html, baseURL)
	if err != nil {
		outgoingLinks = nil
	}

	imageUrls, err := getImageUrlsFromHTML(html, baseURL)
	if err != nil {
		imageUrls = nil
	}

	pageData.OutgoingLinks = outgoingLinks
	pageData.ImageUrls = imageUrls
	return pageData
}

type Parser interface {
	GetFirstElement(html, element string) string
	GetFirstText(html, element string) string
	FindUrls(baseURL *url.URL, html, element, attribute string) ([]string, error)
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

func getUrlsFromHTML(htmlBody string, baseURL *url.URL) ([]string, error) {
	parser := goquery.Parser{}
	urls, err := parser.FindUrls(baseURL, htmlBody, "a", "href")
	if err != nil {
		return nil, err
	} else {
		return urls, nil
	}
}

func getImageUrlsFromHTML(htmlBody string, baseURL *url.URL) ([]string, error) {
	parser := goquery.Parser{}
	urls, err := parser.FindUrls(baseURL, htmlBody, "img", "src")
	if err != nil {
		return nil, err
	} else {
		return urls, nil
	}
}

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
