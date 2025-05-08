package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/eslami200117/ala_unlimited/model/request"
	"net/http"
	"os"
	"time"

	"github.com/eslami200117/ala_unlimited/config"
	"github.com/eslami200117/ala_unlimited/model/extract"

	"github.com/rs/zerolog"
)

type Core struct {
	conf        *config.Config
	Q           *FixedQueue
	notif       chan string
	messageResp chan string
	sellerMap   map[int]string
	logger      zerolog.Logger
	reqChan     chan request.Request
	resChan     chan *extract.ExtProductPrice
}

var sellerHard = map[int]string{
	1105946: "تک ترند",
	1720400: "پاورتک شاپ",
}

func NewCore(cnf *config.Config) *Core {
	_logger := zerolog.New(os.Stderr).
		With().Str("package", "service").
		Caller().Timestamp().Logger()

	ntf := Core{
		conf:        cnf,
		Q:           NewFixedQueue(100),
		notif:       make(chan string),
		messageResp: make(chan string),
		sellerMap:   sellerHard,
		logger:      _logger,
	}

	return &ntf

}

func (c *Core) Start(maxRate int, duration int) error {
	c.reqChan = make(chan request.Request, 100)
	c.resChan = make(chan *extract.ExtProductPrice, 100)
	runTicker := time.NewTicker(time.Duration(maxRate) * time.Microsecond)
	checkTicker := time.NewTicker(c.conf.CheckInterval * time.Minute)
	timeout := time.After(time.Duration(duration) * time.Minute)

	go c.run(runTicker)
	go c.manager(timeout, checkTicker)

	return nil
}

func (c *Core) manager(timeout <-chan time.Time, checkTicker *time.Ticker) {
	defer checkTicker.Stop()
	for {
		select {
		case <-checkTicker.C:
			c.notif <- "log"
			err := c.SendTelegramMessage(<-c.messageResp)
			if err != nil {
				c.logger.Error().
					Err(err).
					Msg("failed to send telegram message")
			}

		case <-timeout:
			c.notif <- "done"
			err := c.SendTelegramMessage(<-c.messageResp)
			if err != nil {
				c.logger.Error().
					Err(err).
					Msg("failed to send telegram message")
			}
			return
		}
	}
}

func (c *Core) run(ticker *time.Ticker) {
	number := 0
	Err := 0
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			productPrice := &extract.ExtProductPrice{}
			number++
			req, ok := <-c.reqChan
			if !ok {
				c.logger.Info().Msg("req Channel closed by client")
				break
			}
			dkp := req.DKP
			colors := req.Colors

			url := fmt.Sprintf(c.conf.DigiKalaAPIURL, dkp)
			resp, err := http.Get(url)
			if err != nil {
				Err++
				c.logger.Error().
					Err(err).
					Str("dkp", dkp).
					Msg("failed to request digikala")

			} else {
				productPrice, err = c.findPrice(colors, resp)
				if err != nil {
					c.logger.Error().
						Err(err).
						Str("dkp", dkp).
						Msg("failed to extract data")
				}
			}

			productPrice.Status = resp.StatusCode
			c.resChan <- productPrice

		case req := <-c.notif:
			if req == "done" {
				close(c.resChan)
				c.messageResp <- "quit successfully!"
				return
			} else if req == "log" {
				c.messageResp <- fmt.Sprintf("number of Request %d, number of Error %d", number, Err)
				number = 0
				Err = 0
			}
		}
	}
}

func (c *Core) SendTelegramMessage(message string) error {
	c.logger.Info().Msg(fmt.Sprintf("[telegram] %s...", message[:7]))
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
	resp.Body.Close()
	return nil
}
