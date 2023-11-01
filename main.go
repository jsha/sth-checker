// This is a program to check the performance of a CT log's get-sth-consistency
// endpoint for two categories of parameters: "old" (near the left hand side of
// the tree) and "new" (near the right hand side of the tree).
//
// Once per second, it fires off two goroutines. The "old" goroutine picks a random
// of length 300 within the first 10,000 entries in the log and fetches get-sth-consistency.
//
// The "new" goroutine fetches get-sth, parses the tree_size value, and then calls
// get-sth-consistency with second = tree_size - 1 and first = tree_size - 300.
//
// Each goroutine prints a summary of the request, including the start and end time,
// the duration, and any error.
//
// Because a new goroutine is fired off for each case, when one response is slow
// it doesn't prevent new responses from being started.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	logURL := flag.String("logURL", "", "URL of the certificate transparency log")
	flag.Parse()
	for {
		go checkOldValue(*logURL)
		go checkNewValue(*logURL)
		time.Sleep(time.Second)
	}
	select {}
}

func checkOldValue(logURL string) {
	first := rand.Intn(10000)
	second := first + 300
	start := time.Now().UTC()
	resp, err := http.Get(fmt.Sprintf("%s/ct/v1/get-sth-consistency?first=%d&second=%d", logURL, first, second))
	end := time.Now().UTC()
	if err == nil && resp.StatusCode != 200 {
		var body []byte
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Print("error reading error body")
		}
		err = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	dur := end.Sub(start).Truncate(time.Millisecond)
	log.Printf("OLD latency: %8s, Start: %s, End: %s, err=%s", dur, start, end, err)
}

func checkNewValue(logURL string) {
	resp, err := http.Get(fmt.Sprintf("%s/ct/v1/get-sth", logURL))
	if err != nil {
		log.Printf("Error getting STH: %s", err)
		return
	}
	var treeSize struct {
		TreeSize int `json:"tree_size"`
	}
	err = json.NewDecoder(resp.Body).Decode(&treeSize)
	if err != nil {
		log.Printf("Error decoding STH: %s", err)
		return
	}
	second := treeSize.TreeSize - 1
	first := second - 300
	start := time.Now().UTC()
	resp, err = http.Get(fmt.Sprintf("%s/ct/v1/get-sth-consistency?first=%d&second=%d", logURL, first, second))
	end := time.Now().UTC()
	if err == nil && resp.StatusCode != 200 {
		var body []byte
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Print("error reading error body")
		}
		err = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	dur := end.Sub(start).Truncate(time.Millisecond)
	log.Printf("NEW latency: %8s, Start: %s, End: %s, err=%s", dur, start, end, err)
}
