package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/sclevine/agouti"
	"golang.org/x/net/websocket"
)

var driver *agouti.WebDriver

func init() {
	driver = agouti.ChromeDriver(
		agouti.ChromeOptions("args", []string{
			"--headless",
			"--window-size=1680,1050",
			"--no-sandbox",
			"--disable-gpu",
		}),
	)
	err := driver.Start()
	if err != nil {
		log.Printf("Failed to start driver. please restart server: %v", err)
		os.Exit(1)
	}
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

func getEnv(key, defaultValue string) string {
	env := os.Getenv(key)
	if env == "" {
		return defaultValue
	}
	return env
}

func index(w http.ResponseWriter, _ *http.Request) {
	webSocketHost := getEnv("ANALYZER_WEBSOCKET_HOST", "localhost:8080")

	params := map[string]string{
		"WebSocketHost": webSocketHost,
	}

	t := template.Must(template.ParseFiles("index.html.tpl"))
	if err := t.ExecuteTemplate(w, "index.html.tpl", params); err != nil {
		log.Printf("Failed to parse template: %v", err)
	}
}

func websocketHandler(ws *websocket.Conn) {
	for {
		var err error
		var url string

		if err = websocket.Message.Receive(ws, &url); err != nil {
			log.Printf("couldn't receive websocket message %v", err)
			break
		}

		_, err = NewHTTPClient().Get(url)
		if err != nil {
			WriteResponse(ws, err.Error(), statusFailure)
			continue
		}

		rawHTML, err := getHTML(url)
		if err != nil {
			WriteResponse(ws, err.Error(), statusFailure)
			continue
		}

		document, err := getDocument(rawHTML)
		if err != nil {
			WriteResponse(ws, err.Error(), statusFailure)
			continue
		}

		analyzer := NewAnalyzer(ws, url, rawHTML, document)
		analyzer.Start()
		analyzer.Wait()
		analyzer.Complete()
	}
}

func main() {
	defer driver.Stop()
	http.HandleFunc("/", index)
	http.Handle("/ws", websocket.Handler(websocketHandler))
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Printf("Failed to start server. please restart server: %v", err)
		os.Exit(1)
	}
}
