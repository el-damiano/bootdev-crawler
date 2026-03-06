package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/el-damiano/bootdev-crawler/internal/goquery"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("no website provided")
		os.Exit(1)
	} else if len(os.Args) > 2 {
		fmt.Println("too many arguments provided")
		os.Exit(1)
	}

	fmt.Printf("starting crawl of: %v\n", os.Args[1])
	html, err := getHTML(os.Args[1])
	if err != nil {
		fmt.Printf("error getting HTML: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\ndata:\n%v\n", html)
}

func getHTML(rawURL string) (string, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "BootCrawler/1.0")
	req.UserAgent()

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode > 399 {
		return "", fmt.Errorf("HTTP ERROR: %v", res.StatusCode)
	}

	contentType := res.Header.Get("content-type")
	if contentType != "text/html" {
		return "", fmt.Errorf("ERROR: unsupported content-type: %v", contentType)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
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
