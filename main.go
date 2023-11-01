// This is a program that takes as a flag the name of a certificate transparency log. It spawns two goroutines.
// One goroutine repeatedly checks the certificate transparency log's get-sth-consistency endpoint for small random
// values of the `first` and `second` parameters. Call this the "old values" goroutine.
//
// The second goroutine, called the "new values" goroutine, repeatedly checks the certificate transparency log's
// get-sth endpoint and parses it to find the tree_size value. It then calls the get-sth-consistency endpoint with
// second = tree_size - 1 and first = tree_size - 300.
//
// The checks on both goroutines use a timeout value of 30 seconds. After each check, they print the start time
// of the check, the end time of the check, whether it was successful or not, and how long it took.
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
