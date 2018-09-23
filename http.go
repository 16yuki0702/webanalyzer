package main

import (
	"crypto/tls"
	"net/http"
)

// NewHTTPClient returns new http client.
func NewHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
}
