package main

import (
	"log"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/sclevine/agouti"
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
