package notification

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"time"
)


const (
	cpuThreshold   = 80.0
	ramThreshold   = 80.0
	telegramAPIURL = "https://api.telegram.org/bot%s/sendMessage"
)





type Notifi struct {
	botToken  		string
	chatID    		string	
	Q               *FixedQueue
	notif			chan string
	response 		chan string
	maxRate			int
	duration		int
}

func NewNotifi(botToken, chatID string) *Notifi{
	ntf := Notifi{
		botToken: 	botToken,
		chatID: 	chatID,
		Q: 			NewFixedQueue(100),
		notif:		make(chan string, 1),
		response:	make(chan string),
	}

	return &ntf

}


func (ntf *Notifi) ReConfig(maxRate int, duration int) error {
	ntf.Q.Clear()
	ntf.maxRate = maxRate
	ntf.duration = duration

	return nil
}



func (ntf *Notifi) Start() error {
	dkpFile, err := os.Open("dkp.txt")
	if err != nil {
		panic(err)
	}
	defer dkpFile.Close()

	dkpList := []string{}


	scanner := bufio.NewScanner(dkpFile)
	for scanner.Scan() {
		line := scanner.Text()
		dkpList = append(dkpList, line)
	}

	runTicker := time.NewTicker(time.Duration(ntf.maxRate) * time.Microsecond)
	checkTicker := time.NewTicker(time.Hour * 2) 
	go ntf.run(runTicker, dkpList)

	timeout := time.After(time.Duration(ntf.duration) * time.Minute)

	go func(){
		defer checkTicker.Stop()
		for {
			select {
			case <-checkTicker.C:
				ntf.notif <- "log"
				ntf.SendTelegramMessage(<-ntf.response)
	
			case <-timeout:
				ntf.notif <- "done"
				ntf.SendTelegramMessage(<-ntf.response)
				return
			}
		}
	}()

	return nil
}




func (ntf *Notifi) run(ticker *time.Ticker, dkpList []string) {
	number := 0 
	Err := 0
	defer ticker.Stop()
	for {
		select {
		case <- ticker.C:
			number++
			dkp := dkpList[rand.IntN(len(dkpList))]
			url := fmt.Sprintf("https://api.digikala.com/v2/product/%s", dkp)
			resp, err := http.Get(url)	
			statusCode := 0
			if err != nil {
				Err++
				ntf.Q.Add(fmt.Sprintf("[%d] DKP: %s | Error: %v\n", 1, dkp, err))
			} else {
				statusCode = resp.StatusCode
				resp.Body.Close()
				if statusCode != 200 {
					Err++
				}
			}

		case req := <-ntf.notif:
			if req == "done" {
				ntf.response <- "quite successfully!"
				return
			} else if req == "log" {
				ntf.response <- fmt.Sprintf("number of Request %d, number of Error %d", number, Err)
				number = 0
				Err = 0
			}
		}
	}
}


func (ntf *Notifi) SendTelegramMessage(message string) error {
	log.Println("send message:", message)
	url := fmt.Sprintf(telegramAPIURL, ntf.botToken)

	payload := map[string]string{
		"chat_id": ntf.chatID,
		"text":    message,
	}
	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

