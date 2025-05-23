package service

import (
	"fmt"
	"strconv"
	"strings"
)

func (c *Core) messaging() {
	c.logger.Info().Msg("Messaging service started")
	for {
		select {
		case req := <-c.notif:
			switch req {
			case "done":
				c.Quit()
			case "log":
				c.messageResp <- fmt.Sprintf("number of requests: %d, number of errors: %d", c.requestCount, c.errorCount)
				c.logger.Info().Msgf("number of requests: %d, number of errors: %d", c.requestCount, c.errorCount)
				c.requestCount = 0
				c.errorCount = 0
			case "start":
				text := <-c.notif
				parts := strings.Fields(text)
				if len(parts) != 2 {
					c.messageResp <- "❌ Please send exactly 2 numbers: max rate and duration."
					continue
				}
				maxRate, err1 := strconv.Atoi(parts[0])
				duration, err2 := strconv.Atoi(parts[1])
				if err1 != nil || err2 != nil {
					c.messageResp <- "❌ Invalid input. Send two integers."
					continue
				}
				c.Start(maxRate, duration)
				c.messageResp <- "OK"
			case "set":
				text := <-c.notif
				lines := strings.Split(text, "\n")
				sellers := make(map[int]string)
				for _, line := range lines {
					parts := strings.Fields(line)
					if len(parts) < 2 {
						c.messageResp <- "❌ Invalid input."
						continue
					}
					id, err := strconv.Atoi(parts[0])
					if err != nil {
						c.messageResp <- "❌ Invalid input."
						continue
					}
					sellers[id] = strings.Join(parts[1:], " ")
				}
				c.SetSellers(sellers)
				c.messageResp <- "OK"

			}

		}
	}
}
