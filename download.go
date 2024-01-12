package main

import (
	"io"
	"log"
	"net/http"
)

// Downloads the contents of the url
func download(url string, ch chan DownloadResult) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		// Read the contents of the html body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		ch <- DownloadResult{url, body, nil}
	}
}
