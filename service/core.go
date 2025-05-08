package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/eslami200117/ala_unlimited/model/request"
	pb "github.com/eslami200117/ala_unlimited/protocpb"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/eslami200117/ala_unlimited/config"
	"github.com/eslami200117/ala_unlimited/model/extract"

	"github.com/rs/zerolog"
)

type Core struct {
	pb.UnimplementedPriceServiceServer

	conf        *config.Config
	Q           *FixedQueue
	running     bool
	done        chan struct{}
	notif       chan string
	messageResp chan string
	sellerMap   map[int]string
	logger      zerolog.Logger
	reqQueue    chan request.Request
	resQueue    chan *extract.ExtProductPrice
	sellerMutex sync.RWMutex
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
		logger:      _logger,
		sellerMap:   make(map[int]string),
		sellerMutex: sync.RWMutex{},
	}

	return &ntf

}

func (c *Core) Start(maxRate int, duration int) error {
	c.reqQueue = make(chan request.Request, 100)
	c.resQueue = make(chan *extract.ExtProductPrice, 100)
	c.running = true
	runTicker := time.NewTicker(time.Duration(maxRate) * time.Microsecond)
	checkTicker := time.NewTicker(c.conf.CheckInterval * time.Minute)
	timeout := time.After(time.Duration(duration) * time.Minute)
	go func() {
		<-timeout
		if c.running {
			c.done <- struct{}{}
			close(c.done)
		}
	}()

	go c.run(runTicker)
	go c.manager(checkTicker)

	return nil
}

func (c *Core) manager(checkTicker *time.Ticker) {
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

		case <-c.done:
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
			req, ok := <-c.reqQueue
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
			c.resQueue <- productPrice

		case req := <-c.notif:
			if req == "done" {
				close(c.resQueue)
				c.running = false
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

func (c *Core) Quit() {
	if c.running {
		c.done <- struct{}{}
		c.logger.Info().Msg("quit successfully")
	} else {
		c.logger.Info().Msg("try to quit when it's already stopped")
		c.messageResp <- "already stopped"
	}
}

func (c *Core) SetSellers(sellers map[int]string) {
	c.sellerMutex.Lock()
	defer c.sellerMutex.Unlock()
	for id, name := range sellers {
		c.sellerMap[id] = name
	}
}
