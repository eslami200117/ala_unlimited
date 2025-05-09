package service

import (
	"bytes"
	"context"
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

	conf    *config.Config
	Q       *FixedQueue
	running bool
	//done        chan struct{}
	notif       chan string
	messageResp chan string
	sellerMap   map[int]string
	logger      zerolog.Logger
	reqQueue    chan request.Request
	resQueue    chan *extract.ExtProductPrice
	sellerMutex sync.RWMutex
	cancel      context.CancelFunc
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

func (c *Core) Start(maxRate, duration int) error {
	c.reqQueue = make(chan request.Request, 64)
	c.resQueue = make(chan *extract.ExtProductPrice, 64)
	c.running = true

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(duration)*time.Minute)
	c.cancel = cancel

	runTicker := time.NewTicker(time.Duration(maxRate) * time.Microsecond)
	checkTicker := time.NewTicker(c.conf.CheckInterval * time.Minute)

	go c.run(ctx, runTicker)
	go c.manager(ctx, checkTicker)

	return nil
}

func (c *Core) manager(ctx context.Context, checkTicker *time.Ticker) {
	defer checkTicker.Stop()

	for {
		select {
		case <-checkTicker.C:
			c.notif <- "log"
			msg := <-c.messageResp
			if err := c.SendTelegramMessage(msg); err != nil {
				c.logger.Error().
					Err(err).
					Msg("failed to send telegram message")
			}

		case <-ctx.Done():
			c.notif <- "done"
			msg := <-c.messageResp
			if err := c.SendTelegramMessage(msg); err != nil {
				c.logger.Error().
					Err(err).
					Msg("failed to send telegram message")
			}
			return
		}
	}
}

func (c *Core) run(ctx context.Context, ticker *time.Ticker) {
	var (
		requestCount = 0
		errorCount   = 0
	)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			req, ok := <-c.reqQueue
			if !ok {
				c.logger.Info().Msg("request channel closed")
				return
			}

			requestCount++
			productPrice := &extract.ExtProductPrice{}

			url := fmt.Sprintf(c.conf.DigiKalaAPIURL, req.DKP)
			resp, err := http.Get(url)
			if err != nil {
				errorCount++
				c.logger.Error().
					Err(err).
					Str("dkp", req.DKP).
					Msg("failed to request digikala")
			} else {
				var extractErr error
				productPrice, extractErr = c.findPrice(req.Colors, resp)
				if extractErr != nil {
					c.logger.Error().
						Err(extractErr).
						Str("dkp", req.DKP).
						Msg("failed to extract data")
				}
				productPrice.Status = resp.StatusCode
			}

			// Always send a response regardless of error state
			c.resQueue <- productPrice

		case req := <-c.notif:
			switch req {
			case "done":
				close(c.resQueue)
				c.running = false
				c.messageResp <- "quit successfully!"
				return
			case "log":
				c.messageResp <- fmt.Sprintf("number of requests: %d, number of errors: %d", requestCount, errorCount)
				requestCount = 0
				errorCount = 0
			}

		case <-ctx.Done():
			close(c.resQueue)
			c.running = false
			return
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
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			c.logger.Error().
				Err(err).
				Msg("failed to close response body")
		}
	}()
	return nil
}

func (c *Core) Quit() {
	if c.running {
		c.cancel()
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
