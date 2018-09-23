package main

import (
	"crypto/tls"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/sclevine/agouti"
)

var driver *agouti.WebDriver

func init() {
	driver = agouti.ChromeDriver(
		agouti.ChromeOptions("args", []string{
			"--headless",
			"--window-size=1280,800",
		}),
	)
	err := driver.Start()
	if err != nil {
		log.Printf("Failed to start driver. please restart server: %v", err)
		os.Exit(1)
	}
}

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

func NewHttpClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
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
	a.doc.Find("h1").Each(func(_ int, _ *goquery.Selection) { a.h1 += 1 })
}

func (a *Analyzer) findH2() {
	a.doc.Find("h2").Each(func(_ int, _ *goquery.Selection) { a.h2 += 1 })
}

func (a *Analyzer) findH3() {
	a.doc.Find("h3").Each(func(_ int, _ *goquery.Selection) { a.h3 += 1 })
}

func (a *Analyzer) findH4() {
	a.doc.Find("h4").Each(func(_ int, _ *goquery.Selection) { a.h4 += 1 })
}

func (a *Analyzer) findH5() {
	a.doc.Find("h5").Each(func(_ int, _ *goquery.Selection) { a.h5 += 1 })
}

func (a *Analyzer) findH6() {
	a.doc.Find("h6").Each(func(_ int, _ *goquery.Selection) { a.h6 += 1 })
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
					a.internalLink += 1
					a.pararel(func() {
						_, err := a.httpClient.Get(reqURL)
						if err != nil {
							a.invalidInternalLink += 1
						}
					})
				} else {
					a.externalLink += 1
					a.pararel(func() {
						_, err := a.httpClient.Get(link)
						if err != nil {
							a.invalidExternalLink += 1
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

func getHTML(url string) (string, error) {
	page, err := driver.NewPage(agouti.Browser("chrome"))
	if err != nil {
		return "", errors.Wrap(err, "Failed to open page")
	}

	err = page.Navigate(url)
	if err != nil {
		return "", errors.Wrap(err, "Failed to Navigate")
	}

	content, err := page.HTML()
	if err != nil {
		return "", errors.Wrap(err, "Failed to get html")
	}

	return content, nil
}

func getDocument(html string) (*goquery.Document, error) {
	reader := strings.NewReader(html)

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get document")
	}

	return doc, nil
}

func indexPage() string {
	return `
<html>
	<head>
		<title>analyze web</title>
	</head>
	<body>
		<form action="" method=post>
			<input type=text name=value_post value="">
			<input type=submit name=submit value=submit>
		</form>
	</body>
</html>
`
}

func errorPage(err error) string {
	return `
<html>
    <head>
		<title>analyze web</title>
    </head>
    <body>
		<form action="" method=post>
			<input type=text name=value_post value="">
			<input type=submit name=submit value=submit>
		</form>
		<div>` + html.EscapeString(err.Error()) + `</div>
    </body>
</html>
`
}

func Analyze(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("value_post")
	if url == "" {
		fmt.Fprintf(w, "%s", indexPage())
		return
	}

	httpClient := NewHttpClient()

	_, err := httpClient.Get(url)
	if err != nil {
		fmt.Fprintf(w, "%s", errorPage(err))
		return
	}

	rawHTML, err := getHTML(url)
	if err != nil {
		fmt.Fprintf(w, "%s", errorPage(err))
		return
	}

	document, err := getDocument(rawHTML)
	if err != nil {
		fmt.Fprintf(w, "%s", errorPage(err))
		return
	}

	analyzer := NewAnalyzer(url, document, rawHTML, httpClient)
	analyzer.start()
	analyzer.wait()

	fmt.Fprintf(w, "%s", analyzer.toResponse())
}

func main() {
	defer driver.Stop()
	http.HandleFunc("/", Analyze)
	http.ListenAndServe(":8080", nil)
}
