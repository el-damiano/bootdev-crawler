package goquery

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Parser struct{}

func (p Parser) GetFirstElement(html, element string) string {
	document, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return ""
	}

	selection := document.Find(element)
	firstElement, err := goquery.OuterHtml(selection.First())
	if err != nil {
		return ""
	}

	return firstElement
}

func (p Parser) GetFirstText(html, element string) string {
	document, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return ""
	}

	selection := document.Find(element)
	firstElement := selection.First().Text()
	return firstElement
}

func (p Parser) FindUrls(baseURL *url.URL, html, element, attribute string) ([]string, error) {
	document, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	var urls []string
	document.Find(element).Each(func(_ int, s *goquery.Selection) {
		link, ok := s.Attr(attribute)
		if !ok {
			return
		}
		link = strings.TrimSpace(link)
		if link == "" {
			return
		}

		linkParsed, err := url.Parse(link)
		if err != nil {
			fmt.Printf("couldn't parse link %q: %v\n", link, err)
			return
		}

		linkResolved := baseURL.ResolveReference(linkParsed)
		urls = append(urls, linkResolved.String())

	})

	return urls, nil
}
