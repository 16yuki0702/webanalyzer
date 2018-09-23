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
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/websocket"
)

// Analyzer represents analyzer of web pages.
type Analyzer struct {
	waitGroup  *sync.WaitGroup
	ws         *websocket.Conn
	requestURL string
	rawHTML    string
	document   *goquery.Document
	httpClient *http.Client

	internalLink        int
	invalidInternalLink int
	externalLink        int
	invalidExternalLink int

	startTime      time.Time
	processingTime time.Duration
}

// NewAnalyzer returns new Analyzer.
func NewAnalyzer(ws *websocket.Conn,
	requestURL string,
	rawHTML string,
	document *goquery.Document) *Analyzer {

	return &Analyzer{
		ws:         ws,
		rawHTML:    rawHTML,
		document:   document,
		requestURL: requestURL,
		waitGroup:  &sync.WaitGroup{},
	}
}

// Start starts analyzing web page.
func (a *Analyzer) Start() {
	a.startTime = time.Now()
	a.pararel(a.findTitle)
	a.pararel(a.findDocType)
	for i := 1; i <= 6; i++ {
		a.pararel(a.findHeading(i))
	}
	a.pararel(a.findLinks)
	a.pararel(a.findLoginForm)
}

// Wait waits until end of analyzing web page.
func (a *Analyzer) Wait() {
	a.waitGroup.Wait()
	a.processingTime = time.Since(a.startTime)
}

// Complete sends response of complete of analyzing web page to client.
func (a *Analyzer) Complete() {
	// only these parts are calculated in the goroutine in the goroutine.
	// so these parts are rendered at the end of all the goroutines.
	WriteResponse(a.ws, fmt.Sprintf("internal link count : %d", a.internalLink), statusSuccess)
	WriteResponse(a.ws, fmt.Sprintf("invalid internal link count : %d", a.invalidInternalLink), statusSuccess)
	WriteResponse(a.ws, fmt.Sprintf("external link count : %d", a.externalLink), statusSuccess)
	WriteResponse(a.ws, fmt.Sprintf("invalid external link count : %d", a.invalidExternalLink), statusSuccess)

	WriteResponse(a.ws, fmt.Sprint("analyze complete"), statusComplete)
	WriteResponse(a.ws, fmt.Sprintf("processing time %s", a.processingTime), statusSuccess)
}

// WriteResponse sends response to client.
func WriteResponse(ws *websocket.Conn, message string, status analyzeResponseStatus) {
	if err := websocket.JSON.Send(ws, analyzeResponse{Result: message, Status: status}); err != nil {
		log.Printf("couldn't send websocket response %v", err)
	}
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

func (a *Analyzer) pararel(f func()) {
	a.waitGroup.Add(1)
	go func() {
		defer a.waitGroup.Done()
		f()
	}()
}

func (a *Analyzer) findDocType() {
	firstline := strings.Split(a.rawHTML, "\n")[0]
	r, _ := regexp.Compile("<!DOCTYPE(.*?)>")
	match := r.FindString(firstline)
	WriteResponse(a.ws, fmt.Sprintf("html version : %s", html.EscapeString(match)), statusSuccess)
}

func (a *Analyzer) findTitle() {
	value := a.document.Find("title").Text()
	WriteResponse(a.ws, fmt.Sprintf("title : %s", html.EscapeString(value)), statusSuccess)
}

func (a *Analyzer) findHeading(level int) func() {
	return func() {
		var value int
		findLevel := fmt.Sprintf("h%d", level)
		a.document.Find(findLevel).Each(func(_ int, _ *goquery.Selection) { value++ })
		WriteResponse(a.ws, fmt.Sprintf("%s count : %d", findLevel, value), statusSuccess)
	}
}

func (a *Analyzer) findLinks() {
	parsedURL, _ := url.ParseRequestURI(a.requestURL)
	httpClient := NewHTTPClient()
	ignoreList := map[string]bool{}

	a.document.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		link, _ := s.Attr("href")
		if !ignoreList[link] {
			ignoreList[link] = true

			uri, err := url.ParseRequestURI(link)
			if err == nil {
				if uri.Host == "" {
					reqURL := fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, link)
					a.internalLink++
					a.pararel(func() {
						_, err := httpClient.Get(reqURL)
						if err != nil {
							a.invalidInternalLink++
						}
					})
				} else {
					a.externalLink++
					a.pararel(func() {
						_, err := httpClient.Get(link)
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
	var loginFound bool
	a.document.Find("form").Each(func(_ int, s *goquery.Selection) {
		action, _ := s.Attr("action")
		if strings.Contains(action, "login") {
			loginFound = true
		}
	})
	WriteResponse(a.ws, fmt.Sprintf("contain login form : %s", strconv.FormatBool(loginFound)), statusSuccess)
}
