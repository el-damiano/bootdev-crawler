package goquery

import (
	"github.com/PuerkitoBio/goquery"
	"strings"
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
