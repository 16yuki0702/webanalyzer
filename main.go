package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/websocket"
)

func index(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "%s", indexPage())
}

func websocketHandler(ws *websocket.Conn) {
	for {
		var err error
		var url string

		if err = websocket.Message.Receive(ws, &url); err != nil {
			log.Printf("couldn't receive websocket message %v", err)
			break
		}

		httpClient := NewHTTPClient()

		_, err = httpClient.Get(url)
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

		analyzer := NewAnalyzer(url, document, rawHTML, httpClient, ws)
		analyzer.start()
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
