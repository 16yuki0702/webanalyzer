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
	"golang.org/x/net/websocket"
)

// Analyzer represents analyzer of web pages.
type Analyzer struct {
	waitGroup  *sync.WaitGroup
	ws         *websocket.Conn
	requestURL string
	rawHTML    string
	doc        *goquery.Document
	httpClient *http.Client
	ignoreList map[string]bool
}

// NewAnalyzer returns new Analyzer.
func NewAnalyzer(requestURL string,
	doc *goquery.Document,
	raw string,
	httpClient *http.Client,
	ws *websocket.Conn) *Analyzer {

	return &Analyzer{
		requestURL: requestURL,
		ws:         ws,
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
	value := strings.Replace(strings.Split(a.rawHTML, "\n")[0], "<head>", "", 1)
	a.writeResponse(fmt.Sprintf("<li>htmlversion : %s</li>", html.EscapeString(value)))
}

func (a *Analyzer) findTitle() {
	value := a.doc.Find("title").Text()
	a.writeResponse(fmt.Sprintf("<li>title : %s</li>", html.EscapeString(value)))
}

func (a *Analyzer) findH1() {
	var value int
	a.doc.Find("h1").Each(func(_ int, _ *goquery.Selection) { value++ })
	a.writeResponse(fmt.Sprintf("<li>h1 count : %d</li>", value))
}

func (a *Analyzer) findH2() {
	var value int
	a.doc.Find("h2").Each(func(_ int, _ *goquery.Selection) { value++ })
	a.writeResponse(fmt.Sprintf("<li>h2 count : %d</li>", value))
}

func (a *Analyzer) findH3() {
	var value int
	a.doc.Find("h3").Each(func(_ int, _ *goquery.Selection) { value++ })
	a.writeResponse(fmt.Sprintf("<li>h3 count : %d</li>", value))
}

func (a *Analyzer) findH4() {
	var value int
	a.doc.Find("h4").Each(func(_ int, _ *goquery.Selection) { value++ })
	a.writeResponse(fmt.Sprintf("<li>h4 count : %d</li>", value))
}

func (a *Analyzer) findH5() {
	var value int
	a.doc.Find("h5").Each(func(_ int, _ *goquery.Selection) { value++ })
	a.writeResponse(fmt.Sprintf("<li>h5 count : %d</li>", value))
}

func (a *Analyzer) findH6() {
	var value int
	a.doc.Find("h6").Each(func(_ int, _ *goquery.Selection) { value++ })
	a.writeResponse(fmt.Sprintf("<li>h6 count : %d</li>", value))
}

func (a *Analyzer) findLinks() {
	parsedURL, _ := url.ParseRequestURI(a.requestURL)
	var internalLink, invalidInternalLink int
	var externalLink, invalidExternalLink int

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
					internalLink++
					_, err := a.httpClient.Get(reqURL)
					if err != nil {
						invalidInternalLink++
					}
				} else {
					externalLink++
					_, err := a.httpClient.Get(link)
					if err != nil {
						invalidExternalLink++
					}
				}
			}
		}
	})

	a.writeResponse(fmt.Sprintf("<li>internal link count : %d</li>", internalLink))
	a.writeResponse(fmt.Sprintf("<li>invalid internal link count : %d</li>", invalidInternalLink))
	a.writeResponse(fmt.Sprintf("<li>external link count : %d</li>", externalLink))
	a.writeResponse(fmt.Sprintf("<li>invalid external link count : %d</li>", invalidExternalLink))
}

func (a *Analyzer) writeResponse(message string) {
	if err := websocket.Message.Send(a.ws, message); err != nil {
		log.Printf("couldn't send websocket response %v", err)
	}
}

func (a *Analyzer) findLoginForm() {
	var loginFound bool
	a.doc.Find("form").Each(func(_ int, s *goquery.Selection) {
		action, _ := s.Attr("action")
		if strings.Contains(action, "login") {
			loginFound = true
		}
	})
	a.writeResponse(fmt.Sprintf("<li>contain login form : %s</li>", html.EscapeString(strconv.FormatBool(loginFound))))
}

func (a *Analyzer) wait() {
	a.waitGroup.Wait()
}

func (a *Analyzer) complete() {
	a.writeResponse(fmt.Sprint("<h2>analyze complete</h2>"))
}
