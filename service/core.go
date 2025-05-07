package service

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"time"

	"github.com/eslami200117/ala_unlimited/config"
	"github.com/eslami200117/ala_unlimited/model/extract"

	"github.com/rs/zerolog"
)

type Core struct {
	conf      *config.Config
	Q         *FixedQueue
	notif     chan string
	response  chan string
	maxRate   int
	duration  int
	sellerMap map[int]string
	logger    zerolog.Logger
}

var sellerHard = map[int]string{
	1105946: "تک ترند",
	1720400: "پاورتک شاپ",
}

func NewCore(cnf *config.Config, _logger zerolog.Logger) *Core {
	ntf := Core{
		conf:      cnf,
		Q:         NewFixedQueue(100),
		notif:     make(chan string, 1),
		response:  make(chan string),
		sellerMap: sellerHard,
		logger:    _logger,
	}

	return &ntf

}

func (c *Core) ReConfig(maxRate int, duration int) error {
	c.Q.Clear()
	c.maxRate = maxRate
	c.duration = duration

	return nil
}

func (c *Core) Start() error {
	dkpFile, err := os.Open("dkp.txt")
	if err != nil {
		panic(err)
	}
	defer dkpFile.Close()

	var dkpList []string

	scanner := bufio.NewScanner(dkpFile)
	for scanner.Scan() {
		line := scanner.Text()
		dkpList = append(dkpList, line)
	}

	runTicker := time.NewTicker(time.Duration(c.maxRate) * time.Microsecond)
	checkTicker := time.NewTicker(time.Minute * 2)
	go c.run(runTicker, dkpList)

	timeout := time.After(time.Duration(c.duration) * time.Minute)

	go func() {
		defer checkTicker.Stop()
		for {
			select {
			case <-checkTicker.C:
				c.notif <- "log"
				err := c.SendTelegramMessage(<-c.response)
				if err != nil {
					c.logger.Error().
						Err(err).
						Msg("failed to send telegram message")
				}

			case <-timeout:
				c.notif <- "done"
				err := c.SendTelegramMessage(<-c.response)
				if err != nil {
					c.logger.Error().
						Err(err).
						Msg("failed to send telegram message")
				}
				return
			}
		}
	}()

	return nil
}

func (c *Core) run(ticker *time.Ticker, dkpList []string) {
	number := 0
	Err := 0
	defer ticker.Stop()
	//for {
	select {
	case <-ticker.C:
		productPrice := &extract.ExtProductPrice{}
		number++
		dkp := dkpList[rand.IntN(len(dkpList))]
		color := []string{"نقره ای", "مشکلی", "طوسی", "استیل"}

		url := fmt.Sprintf(c.conf.DigiKalaAPIURL, dkp)
		resp, err := http.Get(url)
		if err != nil {
			Err++
			c.Q.Add(fmt.Sprintf("[%d] DKP: %s | Error: %v\n", 1, dkp, err))
			c.logger.Error().
				Err(err).
				Msg("failed to request digikala")
		} else {
			productPrice, err = c.findPrice(color, resp)
			if err != nil {
				c.logger.Error().
					Err(err).
					Str("dkp", dkp).
					Msg("failed to extract data")
			}
			productPrice.Status = resp.StatusCode

		}

	case req := <-c.notif:
		if req == "done" {
			c.response <- "quit successfully!"
			return
		} else if req == "log" {
			c.response <- fmt.Sprintf("number of Request %d, number of Error %d", number, Err)
			number = 0
			Err = 0
		}
	}
	//}
}

func (c *Core) SendTelegramMessage(message string) error {
	c.logger.Info()
	url := fmt.Sprintf(c.conf.TelegramAPIURL, c.conf.TelegramBotToken)

	payload := map[string]string{
		"chat_id": c.conf.TelegramChatID,
		"text":    message,
	}
	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		c.logger.Error().
			Err(err).
			Msg("failed to send telegram message")
		return err
	}
	defer resp.Body.Close()
	return nil
}
