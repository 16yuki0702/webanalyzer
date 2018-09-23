package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

// Analyzer represents analyzer of web pages.
type Analyzer struct {
	waitGroup  *sync.WaitGroup
	requestURL string
	rawHTML    string
	doc        *goquery.Document
	httpClient *http.Client

	htmlVersion         string
	title               string
	h1                  int
	h2                  int
	h3                  int
	h4                  int
	h5                  int
	h6                  int
	internalLink        int
	invalidInternalLink int
	externalLink        int
	invalidExternalLink int
	hasLoginFrom        bool

	ignoreList map[string]bool
}

// NewAnalyzer returns new Analyzer.
func NewAnalyzer(requestURL string, doc *goquery.Document, raw string, httpClient *http.Client) *Analyzer {
	return &Analyzer{
		requestURL: requestURL,
		doc:        doc,
		rawHTML:    raw,
		httpClient: httpClient,
		ignoreList: map[string]bool{},
		waitGroup:  &sync.WaitGroup{},
	}
}

func (a *Analyzer) start() {
	a.pararel(a.findTitle)
	a.pararel(a.findDocType)
	a.pararel(a.findH1)
	a.pararel(a.findH2)
	a.pararel(a.findH3)
	a.pararel(a.findH4)
	a.pararel(a.findH5)
	a.pararel(a.findH6)
	a.pararel(a.findLinks)
	a.pararel(a.findLoginForm)
}

func (a *Analyzer) pararel(f func()) {
	a.waitGroup.Add(1)
	go func() {
		defer a.waitGroup.Done()
		f()
	}()
}

func (a *Analyzer) findDocType() {
	a.htmlVersion = strings.Replace(strings.Split(a.rawHTML, "\n")[0], "<head>", "", 1)
}

func (a *Analyzer) findTitle() {
	a.title = a.doc.Find("title").Text()
}

func (a *Analyzer) findH1() {
	a.doc.Find("h1").Each(func(_ int, _ *goquery.Selection) { a.h1++ })
}

func (a *Analyzer) findH2() {
	a.doc.Find("h2").Each(func(_ int, _ *goquery.Selection) { a.h2++ })
}

func (a *Analyzer) findH3() {
	a.doc.Find("h3").Each(func(_ int, _ *goquery.Selection) { a.h3++ })
}

func (a *Analyzer) findH4() {
	a.doc.Find("h4").Each(func(_ int, _ *goquery.Selection) { a.h4++ })
}

func (a *Analyzer) findH5() {
	a.doc.Find("h5").Each(func(_ int, _ *goquery.Selection) { a.h5++ })
}

func (a *Analyzer) findH6() {
	a.doc.Find("h6").Each(func(_ int, _ *goquery.Selection) { a.h6++ })
}

func (a *Analyzer) findLinks() {
	parsedURL, _ := url.ParseRequestURI(a.requestURL)

	a.doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		link, _ := s.Attr("href")
		if !a.ignoreList[link] {
			a.ignoreList[link] = true

			uri, err := url.ParseRequestURI(link)
			if err != nil {
				log.Printf("parse error %v", err)
			} else {
				if uri.Host == "" {
					reqURL := fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, link)
					a.internalLink++
					a.pararel(func() {
						_, err := a.httpClient.Get(reqURL)
						if err != nil {
							a.invalidInternalLink++
						}
					})
				} else {
					a.externalLink++
					a.pararel(func() {
						_, err := a.httpClient.Get(link)
						if err != nil {
							a.invalidExternalLink++
						}
					})
				}
			}
		}
	})
}

func (a *Analyzer) findLoginForm() {
	a.doc.Find("form").Each(func(_ int, s *goquery.Selection) {
		action, _ := s.Attr("action")
		if strings.Contains(action, "login") {
			a.hasLoginFrom = true
		}
	})
}

func (a *Analyzer) wait() {
	a.waitGroup.Wait()
}

func (a *Analyzer) toResponse() string {
	return `
<html>
    <head>
		<title>analyze web</title>
    </head>
    <body>
		<form action="" method=post>
			<input type=text name=value_post value=` + html.EscapeString(a.requestURL) + `>
			<input type=submit name=submit value=submit>
		</form>
		<ul>
			<li>html version : ` + html.EscapeString(a.htmlVersion) + `</li>
			<li>title : ` + html.EscapeString(a.title) + `</li>
			<li>h1 count : ` + html.EscapeString(strconv.Itoa(a.h1)) + `</li>
			<li>h2 count : ` + html.EscapeString(strconv.Itoa(a.h2)) + `</li>
			<li>h3 count : ` + html.EscapeString(strconv.Itoa(a.h3)) + `</li>
			<li>h4 count : ` + html.EscapeString(strconv.Itoa(a.h4)) + `</li>
			<li>h5 count : ` + html.EscapeString(strconv.Itoa(a.h5)) + `</li>
			<li>h6 count : ` + html.EscapeString(strconv.Itoa(a.h6)) + `</li>
			<li>internal link count : ` + html.EscapeString(strconv.Itoa(a.internalLink)) + `</li>
			<li>invalid internal link count : ` + html.EscapeString(strconv.Itoa(a.invalidInternalLink)) + `</li>
			<li>external link count : ` + html.EscapeString(strconv.Itoa(a.externalLink)) + `</li>
			<li>invalid external link count : ` + html.EscapeString(strconv.Itoa(a.invalidExternalLink)) + `</li>
			<li>contain login form : ` + html.EscapeString(strconv.FormatBool(a.hasLoginFrom)) + `</li>
		</ul>
    </body>
</html>
`
}
