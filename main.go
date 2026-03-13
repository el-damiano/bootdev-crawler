package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/el-damiano/bootdev-crawler/internal/goquery"
)

type config struct {
	pages              map[string]PageData
	urlBase            *url.URL
	mutex              *sync.Mutex
	concurrencyControl chan struct{}
	waitGroup          *sync.WaitGroup
	maxPages           int
}

func main() {
	possibleArguments := []string{"<URL>", "<max concurrency>", "<max pages>"}
	arguments := os.Args[1:]

	if len(arguments) < 1 {
		fmt.Println("no website provided")
		fmt.Printf("Usage: bootdev-crawler %v\n", possibleArguments)
		os.Exit(1)
	} else if len(arguments) > len(possibleArguments) {
		fmt.Println("too many arguments provided")
		fmt.Printf("Usage: bootdev-crawler %v\n", possibleArguments)
		os.Exit(1)
	}

	urlBaseRaw := os.Args[1]
	urlBase, err := url.Parse(urlBaseRaw)
	if err != nil {
		fmt.Printf("ERROR(crawl): couldn't parse '%s': %v\n", urlBaseRaw, err)
		return
	}

	fmt.Println(">>> START of crawling")

	bufferSize := 8
	if len(arguments) == 2 {
		argSize, _ := strconv.Atoi(os.Args[2])
		bufferSize = argSize
	}

	maxPages := 100
	if len(arguments) == 3 {
		argPages, _ := strconv.Atoi(os.Args[3])
		maxPages = argPages
	}

	config := config{
		pages:              make(map[string]PageData),
		urlBase:            urlBase,
		concurrencyControl: make(chan struct{}, bufferSize),
		mutex:              &sync.Mutex{},
		waitGroup:          &sync.WaitGroup{},
		maxPages:           maxPages,
	}

	config.waitGroup.Add(1)
	config.crawlPage(urlBaseRaw)
	config.waitGroup.Wait()

	for url, pageData := range config.pages {
		fmt.Printf("%s - %s\n", url, strings.TrimSpace(pageData.Heading))
	}

}

func (config *config) crawlPage(urlCurrentRaw string) {
	config.concurrencyControl <- struct{}{}
	defer func() {
		<-config.concurrencyControl
		config.waitGroup.Done()
	}()
	if config.maxPageCountExceeded() {
		return
	}

	urlParsed, err := url.Parse(urlCurrentRaw)
	if err != nil {
		fmt.Printf("ERROR: couldn't parse '%s': %v\n", urlCurrentRaw, err)
		return
	}

	if config.urlBase.Hostname() != urlParsed.Hostname() {
		return
	}

	urlNormalized, err := urlNormalize(urlCurrentRaw)
	if err != nil {
		fmt.Printf("ERROR: couldn't normalize '%s': %v\n", urlCurrentRaw, err)
		return
	}

	visitedBefore := config.addPageVisit(urlNormalized)
	if visitedBefore {
		return
	}
	fmt.Printf(">>> crawling: %v\n", urlCurrentRaw)

	htmlRaw, err := getHTML(urlCurrentRaw)
	if err != nil {
		fmt.Printf("ERROR: couldn't get HTML of '%s': %v\n", urlCurrentRaw, err)
	}

	pageData := extractPageData(htmlRaw, urlCurrentRaw)
	config.setPageData(urlNormalized, pageData)

	for _, urlNext := range pageData.OutgoingLinks {
		config.waitGroup.Add(1)
		go config.crawlPage(urlNext)
	}

}

func (config *config) addPageVisit(urlNormalized string) (visitedBefore bool) {
	config.mutex.Lock()
	defer config.mutex.Unlock()

	_, visited := config.pages[urlNormalized]
	if visited {
		return true
	} else {
		config.pages[urlNormalized] = PageData{URL: urlNormalized}
		return false
	}
}

func (config *config) maxPageCountExceeded() bool {
	config.mutex.Lock()
	defer config.mutex.Unlock()

	if len(config.pages) > config.maxPages {
		return true
	} else {
		return false
	}
}

func (config *config) setPageData(urlNormalized string, pageData PageData) {
	config.mutex.Lock()
	defer config.mutex.Unlock()
	config.pages[urlNormalized] = pageData
}

// Normalizes a URL.
//
// The URL may be relative or absolute, but not empty.
// It can also be a sentence, as that's not being checked.
func urlNormalize(urlRaw string) (string, error) {
	if urlRaw == "" {
		return "", errors.New("Empty URL")
	}

	urlParsed, err := url.Parse(urlRaw)
	if err != nil {
		return "", fmt.Errorf("couldn't parse URL: %w", err)
	}

	urlFullPath := strings.ToLower(urlParsed.Host + strings.ReplaceAll(urlParsed.Path, "//", "/"))
	urlNormalized := strings.TrimSuffix(urlFullPath, "/")

	return urlNormalized, nil
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
	contentTypeIsUnsupported := !strings.Contains(contentType, "text/html")
	if contentTypeIsUnsupported {
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
