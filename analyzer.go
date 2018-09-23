package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"regexp"
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

type analyzeResponse struct {
	Result string
	Status analyzeResponseStatus
}

type analyzeResponseStatus int

const (
	statusSuccess analyzeResponseStatus = iota
	statusFailure
	statusComplete
)

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
	for i := 1; i <= 6; i++ {
		a.pararel(a.findHeading(i))
	}
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
	r, _ := regexp.Compile("<!DOCTYPE(.*?)>")
	firstline := strings.Split(a.rawHTML, "\n")[0]
	match := r.FindString(firstline)
	writeResponse(a.ws, fmt.Sprintf("html version : %s", html.EscapeString(match)), statusSuccess)
}

func (a *Analyzer) findTitle() {
	value := a.doc.Find("title").Text()
	writeResponse(a.ws, fmt.Sprintf("title : %s", html.EscapeString(value)), statusSuccess)
}

func (a *Analyzer) findHeading(level int) func() {
	return func() {
		var value int
		findLevel := fmt.Sprintf("h%d", level)
		a.doc.Find(findLevel).Each(func(_ int, _ *goquery.Selection) { value++ })
		writeResponse(a.ws, fmt.Sprintf("%s count : %d", findLevel, value), statusSuccess)
	}
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

	writeResponse(a.ws, fmt.Sprintf("internal link count : %d", internalLink), statusSuccess)
	writeResponse(a.ws, fmt.Sprintf("invalid internal link count : %d", invalidInternalLink), statusSuccess)
	writeResponse(a.ws, fmt.Sprintf("external link count : %d", externalLink), statusSuccess)
	writeResponse(a.ws, fmt.Sprintf("invalid external link count : %d", invalidExternalLink), statusSuccess)
}

func (a *Analyzer) findLoginForm() {
	var loginFound bool
	a.doc.Find("form").Each(func(_ int, s *goquery.Selection) {
		action, _ := s.Attr("action")
		if strings.Contains(action, "login") {
			loginFound = true
		}
	})
	writeResponse(a.ws, fmt.Sprintf("contain login form : %s", strconv.FormatBool(loginFound)), statusSuccess)
}

func (a *Analyzer) wait() {
	a.waitGroup.Wait()
}

func (a *Analyzer) complete() {
	writeResponse(a.ws, fmt.Sprint("analyze complete"), statusComplete)
}

func writeResponse(ws *websocket.Conn, message string, status analyzeResponseStatus) {
	if err := websocket.JSON.Send(ws, analyzeResponse{Result: message, Status: status}); err != nil {
		log.Printf("couldn't send websocket response %v", err)
	}
}
