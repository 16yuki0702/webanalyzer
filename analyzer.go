package main

import (
	"fmt"
	"html"
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

	internalLink int
	externalLink int

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
	ResponseComplete(a.ws, fmt.Sprint("analyze complete"))
	ResponseComplete(a.ws, fmt.Sprintf("processing time %s", a.processingTime))
}

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
	ResponseSuccess(a.ws, fmt.Sprintf("html version : %s", html.EscapeString(match)))
}

func (a *Analyzer) findTitle() {
	value := a.document.Find("title").Text()
	ResponseSuccess(a.ws, fmt.Sprintf("title : %s", html.EscapeString(value)))
}

func (a *Analyzer) findHeading(level int) func() {
	return func() {
		var value int
		findLevel := fmt.Sprintf("h%d", level)
		a.document.Find(findLevel).Each(func(_ int, _ *goquery.Selection) { value++ })
		ResponseSuccess(a.ws, fmt.Sprintf("%s count : %d", findLevel, value))
	}
}

func (a *Analyzer) findLinks() {
	ignoreList := map[string]bool{}

	a.document.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		link, _ := s.Attr("href")
		if ignoreList[link] {
			return
		}

		ignoreList[link] = true

		parsedURL, err := url.ParseRequestURI(link)
		if err != nil {
			return
		}

		if parsedURL.Host == "" {
			a.internalLink++
		} else {
			a.externalLink++
		}
	})

	ResponseSuccess(a.ws, fmt.Sprintf("internal link count : %d", a.internalLink))
	ResponseSuccess(a.ws, fmt.Sprintf("external link count : %d", a.externalLink))
}

func (a *Analyzer) findLoginForm() {
	var loginFound bool
	a.document.Find("form").Each(func(_ int, s *goquery.Selection) {
		action, _ := s.Attr("action")
		if strings.Contains(action, "login") {
			loginFound = true
		}
	})
	ResponseSuccess(a.ws, fmt.Sprintf("contain login form : %s", strconv.FormatBool(loginFound)))
}
