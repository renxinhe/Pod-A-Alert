package main

import (
	"encoding/json"
	"fmt"
	"log"
    "net/http"
    "os"
    "regexp"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/jasonlvhit/gocron"
    "github.com/sendgrid/sendgrid-go"
    "github.com/sendgrid/sendgrid-go/helpers/mail"
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

type Alert struct {
	AlertInitTime	time.Time 	
	PrinterSerial	int 		`form:"serial" binding:"required"`
	ReceiverName	string 		`form:"name"`
	ReceiverEmail	string 		`form:"email" binding:"required"`
	AlertSendTime	time.Time
	PrintName		string
	PrintProgress	float64
	PrintState		string
	ShouldEmail		bool
	SentEmail 		bool
}

var alerts []Alert
var alertsBusy bool

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func contains(arr [12]string, str string) bool {
   for _, a := range arr {
      if a == str {
         return true
      }
   }
   return false
}

func setUpServer() {
    router := gin.Default()
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")
	router.GET("/", homePageHandler)
	router.GET("/alerts", getAlertsHandler)
	router.POST("/new_alert", newAlertHandler)
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}
	router.Run(":" + port)
}

func homePageHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl.html", nil)
}

func getAlertsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, alerts)
}

func newAlertHandler(c *gin.Context) {
	if len(alerts) >= 30 {
		c.String(http.StatusBadRequest, "Maximum number of alerts reached.")
		return
	}
	var newAlert Alert
	if c.Bind(&newAlert) == nil {
		matched, err := regexp.MatchString("10\\d\\d\\d", c.PostForm("serial"))
		check(err)
		allSerial := [12]string{"10025", "10026", "10034", "10035", "10038", "10047", "10094", "10097", "10098", "10213", "10453", "10454"}
		if !matched || !contains(allSerial, c.PostForm("serial")) {
			c.String(http.StatusBadRequest, "Bad printer number. The printer number should be 5-digits long, and follows the pattern \"10***\".\nExisting printers: 10025, 10026, 10034, 10035, 10038, 10047, 10094, 10097, 10098, 10213, 10453, 10454.\n")
			return
		}
		emailRegexPattern := "(^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\\.[a-zA-Z0-9-.]+$)"
		matched, err = regexp.MatchString(emailRegexPattern, c.PostForm("email"))
		check(err)
		if !matched {
			c.String(http.StatusBadRequest, "Bad email format.")
			return
		}

		fmt.Printf("Received new alert request for printer number %d. Created by %s. Destination email is %s.\n",
			newAlert.PrinterSerial,
			newAlert.ReceiverName,
			newAlert.ReceiverEmail)
		newAlert.AlertInitTime = time.Now()
		newAlert.ShouldEmail = false
		newAlert.SentEmail = false
		alerts = append(alerts, newAlert)
		c.Redirect(http.StatusMovedPermanently, "/")
	} else {
		c.String(http.StatusBadRequest, "rip")
	}
}

func cronTask() {
	if alertsBusy {
		return
	}
	alertsBusy = true
	for i, alert := range alerts {
		fmt.Printf("Alert %d: [%.2f%%] printer %d with job name %s to %s\n",
			i,
			alert.PrintProgress,
			alert.PrinterSerial,
			alert.PrintName,
			alert.ReceiverEmail)
		updateAlertJobData(&alert)
		if alert.ShouldEmail && !alert.SentEmail && alert.PrintProgress == 100 {
			sendAlertEmail(&alert)
		}
		if !alert.AlertSendTime.IsZero() &&
			time.Since(alert.AlertSendTime).Minutes() > 5 {
				alerts = append(alerts[:i], alerts[i+1:]...)
		} else {
			alerts[i] = alert
		}
	}
	alertsBusy = false
}

func updateAlertJobData(alert *Alert) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://pod0vg.eecs.berkeley.edu:3000/api/aprinters/job_data", nil)
	check(err)
	req.Header.Add("serial", strconv.Itoa(alert.PrinterSerial))
	resp, err := client.Do(req)
	check(err)
	defer resp.Body.Close()

	if resp.StatusCode == 200 { // OK
		jobData := new(JobDataJSON)
		json.NewDecoder(resp.Body).Decode(jobData)
		alert.PrintName = jobData.Job.File.Name
		alert.PrintProgress = jobData.Progress.Completion
		alert.PrintState = jobData.State
		if alert.PrintProgress < 98 && !(alert.PrintProgress == 0 && alert.PrintState == ""){
			alert.ShouldEmail = true
		} else if alert.PrintProgress == 100 && alert.AlertSendTime.IsZero() {
			alert.AlertSendTime = time.Now()
		}
	}
}

func sendAlertEmail(alert *Alert) {
    from := mail.NewEmail("Pod Alert", "noreply@jimren.com")
    subject := fmt.Sprintf("[Pod Alert] %s has finished!", alert.PrintName)
    to := mail.NewEmail(alert.ReceiverName, alert.ReceiverEmail)
    emailBody := fmt.Sprintf(`
Hello %s,
	
Your print job %s on printer number %d has finished at %s.

best,

Pod Alert`,
		alert.ReceiverName,
		alert.PrintName,
		alert.PrinterSerial,
		alert.AlertSendTime.Format("2006-01-02 15:04"))
    content := mail.NewContent("text/plain", emailBody)
    m := mail.NewV3MailInit(from, subject, to, content)

    request := sendgrid.GetRequest(os.Getenv("SENDGRID_API_KEY"), "/v3/mail/send", "https://api.sendgrid.com")
    request.Method = "POST"
    request.Body = mail.GetRequestBody(m)
    response, err := sendgrid.API(request)
    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Println(response.StatusCode)
        fmt.Println(response.Body)
        fmt.Println(response.Headers)
        alert.SentEmail = true
    }
}

func main() {
	alertsBusy = false
    println("Starting server...")
    gocron.Every(10).Seconds().Do(cronTask)
    go setUpServer()
    <- gocron.Start()
}
