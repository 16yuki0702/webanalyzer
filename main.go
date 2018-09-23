package main

import (
	"fmt"
	"net/http"
)

// Analyze analyzes given url pages.
func Analyze(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("value_post")
	if url == "" {
		fmt.Fprintf(w, "%s", indexPage())
		return
	}

	httpClient := NewHTTPClient()

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
