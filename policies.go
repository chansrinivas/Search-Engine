package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// Get the url for the robots.txt
func getPolicy(host string) (string, error) {
	robotsURL := "https://www.ucsc.edu/" + "robots.txt"
	resp, err := http.Get(robotsURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// Get the user agent, disallowed urls and the crawl delay
func parseRobotsTxt(txt string) RobotsTxtRules {
	lines := strings.Split(txt, "\n")
	currentUserAgent := "User-agent: *"
	var currentRules RobotsTxtRules
	currentRules.UserAgents = []string{"*"}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		switch key {
		// User agent
		case "User-agent":
			currentUserAgent = value
		// Disallowed urls
		case "Disallow":
			currentRules.Disallow = append(currentRules.Disallow, strings.Replace(value, "*", ".*", 1))
		// Crawl Delay
		case "Crawl-delay":
			crawlDelay, err := strconv.Atoi(value)
			if err == nil {
				currentRules.CrawlDelay = crawlDelay
			}
		}
	}
	fmt.Println(currentUserAgent)
	return currentRules
}
