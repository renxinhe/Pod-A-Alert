package main

import (
	"fmt"
	"log"
	"encoding/json"
    "net/http"
    "regexp"
    "github.com/jasonlvhit/gocron"
)

type JobDataJSON struct {
	Job struct {
		AveragePrintTime   float64 `json:"averagePrintTime"`
		EstimatedPrintTime float64 `json:"estimatedPrintTime"`
		Filament           struct {
			Tool0 struct {
				Length float64 `json:"length"`
				Volume float64 `json:"volume"`
			} `json:"tool0"`
		} `json:"filament"`
		File struct {
			Date   int    `json:"date"`
			Name   string `json:"name"`
			Origin string `json:"origin"`
			Size   int    `json:"size"`
		} `json:"file"`
		LastPrintTime float64 `json:"lastPrintTime"`
	} `json:"job"`
	Progress struct {
		Completion    float64     `json:"completion"`
		Filepos       int         `json:"filepos"`
		PrintTime     int         `json:"printTime"`
		PrintTimeLeft interface{} `json:"printTimeLeft"`
	} `json:"progress"`
	State string `json:"state"`
}

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
		jobData := new(JobDataJSON)
		json.NewDecoder(resp.Body).Decode(jobData)
		fmt.Fprintf(w, "Print name: %s\n", jobData.Job.File.Name)
		fmt.Fprintf(w, "Print progress: %.2f%%\n", jobData.Progress.Completion)
		fmt.Fprintf(w, "Print state: %s\n", jobData.State)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func task() {
	fmt.Println("I am runnning task.")
}

func main() {
    // http.HandleFunc("/", handler)
    println("Starting server...")
    gocron.Every(1).Second().Do(task)
    <- gocron.Start()

     http.HandleFunc("/", podHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
