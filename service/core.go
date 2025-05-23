package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/eslami200117/ala_unlimited/config"
	"github.com/eslami200117/ala_unlimited/model/extract"
	"github.com/eslami200117/ala_unlimited/model/request"
	"github.com/eslami200117/ala_unlimited/pkg/comm"
	pb "github.com/eslami200117/ala_unlimited/protocpb"

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

func NewCore(cnf *config.Config, _reqChn, _resChn chan string) *Core {

	core := &Core{
		conf:        cnf,
		Q:           NewFixedQueue(100),
		notif:       _reqChn,
		messageResp: _resChn,
		logger:      comm.Logger("core"),
		sellerMap:   make(map[int]string),
		sellerMutex: sync.RWMutex{},
		reqQueue:    make(chan request.Request, 128),
		resQueue:    make(chan *extract.ExtProductPrice, 128),
	}

	go core.messaging()
	return core

}

func (c *Core) Start(maxRate, duration int) error {
	c.running = true

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(duration)*time.Minute)
	if duration == 0 {
		ctx, cancel = context.WithCancel(context.Background())
	}
	c.cancel = cancel

	runTicker := time.NewTicker(time.Duration(maxRate) * time.Microsecond)

	go c.run(ctx, runTicker)

	return nil
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
	}
}

func (c *Core) SetSellers(sellers map[int]string) {
	c.sellerMutex.Lock()
	defer c.sellerMutex.Unlock()
	for id, name := range sellers {
		c.sellerMap[id] = name
	}
}
