package main

import (
	"fmt"
	"log"
	"io/ioutil"
    "net/http"
    "regexp"
)

func podHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Pod %s requested!\n", r.URL.Path[1:])

	serial := r.URL.Path[1:]
	matched, err := regexp.MatchString("100\\d\\d", serial)
	if err != nil {
		log.Fatal(err)
	}
	if !matched {
		fmt.Fprintf(w, "Bad printer serial number: %s\n", serial)
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://pod0vg.eecs.berkeley.edu:3000/api/aprinters/job_data", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("serial", serial)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 { // OK
	    bodyBytes, err := ioutil.ReadAll(resp.Body)
	    if err != nil {
			log.Fatal(err)
		}
	    bodyString := string(bodyBytes)
	    fmt.Fprintf(w, "%s\n", bodyString)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {
    // http.HandleFunc("/", handler)
    http.HandleFunc("/", podHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
