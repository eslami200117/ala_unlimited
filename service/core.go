package service

import (
	"context"
	"fmt"
	"github.com/eslami200117/ala_unlimited/model/request"
	pb "github.com/eslami200117/ala_unlimited/protocpb"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/eslami200117/ala_unlimited/config"
	"github.com/eslami200117/ala_unlimited/model/extract"

	"github.com/rs/zerolog"
)

type Core struct {
	pb.UnimplementedPriceServiceServer

	conf         *config.Config
	Q            *FixedQueue
	running      bool
	notif        chan string
	messageResp  chan string
	sellerMap    map[int]string
	logger       zerolog.Logger
	reqQueue     chan request.Request
	resQueue     chan *extract.ExtProductPrice
	sellerMutex  sync.RWMutex
	cancel       context.CancelFunc
	requestCount int
	errorCount   int
}

func NewCore(cnf *config.Config) *Core {
	_logger := zerolog.New(os.Stderr).
		With().Str("package", "service").
		Caller().Timestamp().Logger()

	ntf := &Core{
		conf:        cnf,
		Q:           NewFixedQueue(100),
		notif:       make(chan string),
		messageResp: make(chan string),
		logger:      _logger,
		sellerMap:   make(map[int]string),
		sellerMutex: sync.RWMutex{},
	}

	go ntf.messaging()
	return ntf

}

func (c *Core) Start(maxRate, duration int) error {
	c.reqQueue = make(chan request.Request, 256)
	c.resQueue = make(chan *extract.ExtProductPrice, 256)
	c.running = true

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(duration)*time.Minute)
	if duration == 0 {
		ctx, cancel = context.WithCancel(context.Background())
	}
	c.cancel = cancel

	runTicker := time.NewTicker(time.Duration(maxRate) * time.Microsecond)
	checkTicker := time.NewTicker(c.conf.CheckInterval * time.Minute)

	go c.run(ctx, runTicker)
	go c.manager(ctx, checkTicker)

	return nil
}

func (c *Core) manager(ctx context.Context, checkTicker *time.Ticker) {
	defer checkTicker.Stop()
	c.notif = make(chan string)
	c.messageResp = make(chan string)

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
			c.logger.Info().Msg("manager is stopping")
			return
		}
	}
}

func (c *Core) run(ctx context.Context, ticker *time.Ticker) {
	c.requestCount = 0
	c.errorCount = 0
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			req, ok := <-c.reqQueue
			if !ok {
				c.logger.Info().Msg("request channel closed")
				return
			}

			c.requestCount++
			productPrice := &extract.ExtProductPrice{}

			url := fmt.Sprintf(c.conf.DigiKalaAPIURL, req.DKP)
			resp, err := http.Get(url)
			if err != nil {
				c.errorCount++
				c.logger.Error().
					Err(err).
					Str("dkp", strconv.Itoa(req.DKP)).
					Msg("failed to request digikala")
				continue
			} else {
				var extractErr error
				productPrice, extractErr = c.findPrice(req.Colors, resp)
				if extractErr != nil {
					c.logger.Error().
						Err(extractErr).
						Str("dkp", strconv.Itoa(req.DKP)).
						Msg("failed to extract data")
					continue
				}
				productPrice.Status = resp.StatusCode
			}

			// Always send a response regardless of error state
			c.resQueue <- productPrice

		case <-ctx.Done():
			close(c.resQueue)
			close(c.reqQueue)
			c.running = false
			return
		}
	}
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
