package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
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
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("no website provided")
		os.Exit(1)
	} else if len(os.Args) > 2 {
		fmt.Println("too many arguments provided")
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
	config := config{
		pages:              make(map[string]PageData),
		urlBase:            urlBase,
		concurrencyControl: make(chan struct{}, bufferSize),
		mutex:              &sync.Mutex{},
		waitGroup:          &sync.WaitGroup{},
	}

	config.crawlPage(urlBaseRaw)
	config.waitGroup.Wait()

	for url, pageData := range config.pages {
		fmt.Printf("%s - %s\n", url, pageData.Heading)
	}

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

func (config *config) crawlPage(urlCurrentRaw string) {
	urlCurrent, err := url.Parse(urlCurrentRaw)
	if err != nil {
		fmt.Printf("ERROR(crawl): couldn't parse '%s': %v\n", urlCurrentRaw, err)
		return
	}

	if config.urlBase.Host != urlCurrent.Host {
		return
	}

	fmt.Printf(">>> crawling: %v\n", urlCurrentRaw)

	html, err := getHTML(urlCurrentRaw)
	if err != nil {
		fmt.Printf("ERROR(crawl): couldn't get HTML of '%s': %v\n", urlCurrentRaw, err)
		return
	}

	urls, err := getUrlsFromHTML(html, config.urlBase)
	if err != nil {
		fmt.Printf("ERROR(crawl): couldn't get URLs from HTML: %v of %s\n", err, urlCurrentRaw)
		return
	}

	for _, urlToCrawlRaw := range urls {
		urlToCrawl, err := url.Parse(urlToCrawlRaw)
		if err != nil {
			fmt.Printf("ERROR(crawl): couldn't parse '%s': %v\n", urlToCrawlRaw, err)
			continue
		}

		if config.urlBase.Host != urlToCrawl.Host {
			continue
		}

		urlNormalized, err := urlNormalize(urlToCrawlRaw)
		if err != nil {
			fmt.Printf("ERROR(crawl): couldn't normalize '%s': %v\n", urlToCrawlRaw, err)
			continue
		}

		htmlCrawled, err := getHTML(urlToCrawlRaw)
		if err != nil {
			fmt.Printf("ERROR(crawl): couldn't get HTML of '%s': %v\n", urlToCrawlRaw, err)
			continue
		}

		isFirstTimeCrawling := config.addPageVisit(htmlCrawled, urlNormalized)
		if !isFirstTimeCrawling {
			continue
		}

		config.waitGroup.Add(1)
		go func(url string) {
			defer config.waitGroup.Done()
			config.concurrencyControl <- struct{}{}
			defer func() { <-config.concurrencyControl }()
			config.crawlPage(urlNormalized)
		}(urlToCrawlRaw)
	}
}

func (config *config) addPageVisit(html, urlNormalized string) (isFirst bool) {
	config.mutex.Lock()
	defer config.mutex.Unlock()

	_, ok := config.pages[urlNormalized]
	if ok {
		return false
	} else {
		config.pages[urlNormalized] = extractPageData(html, urlNormalized)
		return true
	}
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
