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

	var elements []string
	document.Find(element).Each(func(_ int, s *goquery.Selection) {
		link, ok := s.Attr(attribute)
		if ok && len(link) > 0 {
			if link[0] == '/' {
				link = fmt.Sprintf("%v%v", baseURL, link)
			}
			elements = append(elements, link)
		}
	})

	return elements, nil
}
