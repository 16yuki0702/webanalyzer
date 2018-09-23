package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/websocket"
)

func index(w http.ResponseWriter, _ *http.Request) {
	t := template.Must(template.ParseFiles("index.html.tpl"))
	if err := t.ExecuteTemplate(w, "index.html.tpl", nil); err != nil {
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
			writeResponse(ws, err.Error(), statusFailure)
			continue
		}

		rawHTML, err := getHTML(url)
		if err != nil {
			writeResponse(ws, err.Error(), statusFailure)
			continue
		}

		document, err := getDocument(rawHTML)
		if err != nil {
			writeResponse(ws, err.Error(), statusFailure)
			continue
		}

		analyzer := NewAnalyzer(ws, url, rawHTML, document)
		analyzer.start()
		analyzer.wait()
		analyzer.complete()
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
