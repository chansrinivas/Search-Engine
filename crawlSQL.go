package main

import (
	"database/sql"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	wordCountMap       map[string]int // Map of url to total word count in doc
	totalDocsInCorpus  int            // Total number of urls being crawled
	docsContainingWord int            // The count of urls that contain the search term
	mu                 sync.Mutex     // Locks for the database
	tableSize          int            // Total urls table size
	ui                 *URLIndex      // Urls
)

// Struct to hold the results received from the download go routine
type DownloadResult struct {
	url  string
	body []byte
	err  error
}

// Struct to hold the results received from the extract go routine
type ExtractResult struct {
	url       string
	words     []string
	hrefs     []string
	title     string
	Images    []string
	ImageSrcs []string
	altWords  []string
}

type RobotsTxtRules struct {
	UserAgents []string
	Disallow   []string
	CrawlDelay int
}

type (
	Sitemap struct {
		Loc string `xml:"loc"`
	}

	SitemapIndex struct {
		Sitemaps []Sitemap `xml:"sitemap"`
	}

	URLIndex struct {
		URLs []Sitemap `xml:"url"`
	}
)

func crawlSQL(db *sql.DB, url string) {
	buildStopWordsMap()
	// Channels for download, extract and clean
	dlInC := make(chan string, 10)
	defer close(dlInC)
	dlOutC := make(chan DownloadResult, 10)
	defer close(dlOutC)
	extractOutC := make(chan ExtractResult, 10)
	defer close(extractOutC)
	quitC := make(chan bool)

	// When the size of the Hits table does not change, crawling is complete
	go func() {
		for {
			time.Sleep(5 * time.Second)
			err := db.QueryRow("SELECT COUNT(*) FROM Hits").Scan(&tableSize)
			if err != nil {
				log.Fatal(err)
			}
			if tableSize == hitsTableSize {
				quitC <- true
				return
			}
		}
	}()

	// Get the user agent, disallowed urls and crawl delay
	body, _ := getPolicy("https://www.ucsc.edu")
	parseRobotsTxt(body)
	go func() {
		// crawling sitemap
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		sitemapBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		si := &SitemapIndex{}
		err = xml.Unmarshal(sitemapBody, &si)
		if err != nil {
			log.Fatal(err)
		}

		// Unmarshal the individual sitemaps
		for _, sitemap := range si.Sitemaps {
			resp, err := http.Get(sitemap.Loc)
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			sitemapBody, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}

			err = xml.Unmarshal(sitemapBody, &ui)
			if err != nil {
				log.Fatal(err)
			}

			for _, sitemap := range ui.URLs {
				url := sitemap.Loc
				url = strings.ReplaceAll(url, "%20", "")
				// Add it to the download in channel
				dlInC <- url
			}
		}
	}()

outer:

	// Select case pattern for concurrency - download, extract, and clean urls
	for {
		select {
		case url := <-dlInC:
			go download(url, dlOutC)
		case dlRes := <-dlOutC:
			go extract(dlRes.url, dlRes.body, extractOutC)
		case cleanRes := <-extractOutC:
			mu.Lock()
			go populateIndex(db, cleanRes.url, cleanRes, dlInC)
			mu.Unlock()
		case <-quitC:
			break outer
		}
	}
}
