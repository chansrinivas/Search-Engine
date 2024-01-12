package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Serve the local host
func serveLocalHost() {
	formFileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/", formFileServer)
	http.Handle("/project06.css", formFileServer)

	fmt.Printf("Server is running on :%d...\n", 8080)
	log.Fatalln(http.ListenAndServe(":8080", nil))
}

// Tables for indexing
func databases() *sql.DB {
	db, err := sql.Open("sqlite3", "index.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE Words (
			Term_ID INTEGER PRIMARY KEY,
			word TEXT UNIQUE
		);
    	`)

	_, err = db.Exec(`
		CREATE TABLE Urls (
			url_ID INTEGER PRIMARY KEY,
			url_name TEXT UNIQUE NOT NULL,
			word_count INT,
			title TEXT
		);
		`)

	_, err = db.Exec(`
		CREATE TABLE Hits (
			Hit_ID INTEGER PRIMARY KEY,
			Term_ID INTEGER,
			url_ID INTEGER,
			Term_Count INTEGER,
			UNIQUE (Term_ID, url_ID),
			FOREIGN KEY (Term_ID) REFERENCES Words(Term_ID),
			FOREIGN KEY (url_ID) REFERENCES Urls(url_ID)
		);
		`)

	_, err = db.Exec(`
		CREATE TABLE BigramHits (
			ID INTEGER PRIMARY KEY,
			Term1_ID INTEGER,
			Term2_ID INTEGER,
			url_ID INTEGER,
			Term_Count INTEGER,
			UNIQUE (Term1_ID, Term2_ID, url_ID),
			FOREIGN KEY (Term1_ID) REFERENCES Words(Term_ID),
			FOREIGN KEY (Term2_ID) REFERENCES Words(Term_ID),
			FOREIGN KEY (url_ID) REFERENCES Urls(url_ID)
		);
		`)

	_, err = db.Exec(`
	CREATE TABLE ImageWords (
		AltWord_ID INTEGER PRIMARY KEY,
		alt_word TEXT UNIQUE
	);
	`)

	_, err = db.Exec(`
	CREATE TABLE ImageUrls (
		id INTEGER PRIMARY KEY,
		url_name TEXT UNIQUE NOT NULL,
		title TEXT,
		altword_count INTEGER
	);
	`)

	_, err = db.Exec(`
	CREATE TABLE ImageHits (
    	ID INTEGER PRIMARY KEY,
		alt_word_id INTEGER,
    	url_ID INTEGER,
    	imgSrc TEXT,
		term_count INTEGER,
   		FOREIGN KEY (url_ID) REFERENCES Urls(id),
    	FOREIGN KEY (alt_word_id) REFERENCES ImageWords(AltWord_ID),
    	UNIQUE (url_ID, alt_word_id, imgSrc)
	);
	`)

	if err != nil {
		log.Fatal(err)
	}
	return db
}

func main() {
	go serveLocalHost()
	db := databases()
	defer db.Close()

	// Recursive sitemap
	url := "https://www.ucsc.edu/wp-sitemap.xml"
	resultTemplate, err := template.ParseFiles("./static/results.html")
	if err != nil {
		log.Fatal(err)
	}
	go SQLSearchEngineHandler(db, url, resultTemplate)
	body, _ := getPolicy(url)
	parseRobotsTxt(body)

	crawlSQL(db, url)

	for {
		time.Sleep(1 * time.Millisecond)
	}
}
