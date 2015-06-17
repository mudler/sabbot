package main

import (
	"github.com/ajanicij/goduckgo/goduckgo"

	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

// UrlTitle attempts to extract the title of the page that a
// pasted URL points to.
// Returns a string message with the title and URL on success, or a string
// with an error message on failure.
// taken from github.com/darthlukan/gobot and adapted, thank you!
func UrlTitle(msg string) string {
	var (
		newMsg, url, title, word string
	)

	regex, _ := regexp.Compile(`(?i)<title>(.*?)<\/title>`)

	msgArray := strings.Split(msg, " ")

	for _, word = range msgArray {
		if strings.Contains(word, "http") {
			url = word
			break
		}
		if !strings.Contains(word, "http") && strings.Contains(word, "www") {
			url = "http://" + word
			break
		}

	}

	resp, err := http.Get(url)

	if err != nil {
		return fmt.Sprintf("Could not resolve URL %v, beware...\n", url)
	}

	defer resp.Body.Close()

	rawBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Could not read response Body of %v ...\n", url)
	}

	body := string(rawBody)
	noNewLines := strings.Replace(body, "\n", "", -1)
	noCarriageReturns := strings.Replace(noNewLines, "\r", "", -1)
	notSoRawBody := noCarriageReturns

	titleMatch := regex.FindStringSubmatch(notSoRawBody)
	if len(titleMatch) > 1 {
		title = strings.TrimSpace(titleMatch[1])
	} else {
		title = fmt.Sprintf("Title Resolution Failure")
	}
	newMsg = fmt.Sprintf("[ %v ]( %v )\n", title, url)

	return newMsg
}

// WebSearch takes a query string as an argument and returns
// a formatted string containing the results from DuckDuckGo.
// taken from github.com/darthlukan/gobot thank you!
func SearchCmd(query string) string {
	query = strings.TrimSpace(query)
	msg, err := goduckgo.Query(query)
	if err != nil {
		return fmt.Sprintf("DDG Error: %v\n", err)
	}

	switch {
	case len(msg.RelatedTopics) > 0:
		return fmt.Sprintf("First Topical Result: [ %s ]( %s )\n", msg.RelatedTopics[0].FirstURL, msg.RelatedTopics[0].Text)
	case len(msg.Results) > 0:
		return fmt.Sprintf("First External result: [ %s ]( %s )\n", msg.Results[0].FirstURL, msg.Results[0].Text)
	case len(msg.Redirect) > 0:
		return fmt.Sprintf("Redirect result: %s\n", UrlTitle(msg.Redirect))
	default:
		return fmt.Sprintf("Query: '%s' returned no results.\n", query)
	}
}
