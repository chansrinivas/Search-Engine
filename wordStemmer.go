package main

import "github.com/kljensen/snowball"

// Function stems each word according to the imported snowball code
func stemWords(word string) string {
	stemmed, err := snowball.Stem(word, "english", true)
	if err == nil {
		return stemmed
	}
	// If there is an error
	return word
}

var hitsTableSize int = 6263
