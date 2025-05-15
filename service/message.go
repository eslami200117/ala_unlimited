package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

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

func (c *Core) messaging() {
	for {
		select {
		case req := <-c.notif:
			switch req {
			case "done":
				close(c.resQueue)
				close(c.reqQueue)
				c.running = false
				c.messageResp <- "quit successfully!"
				return
			case "log":
				c.messageResp <- fmt.Sprintf("number of requests: %d, number of errors: %d", c.requestCount, c.errorCount)
				c.requestCount = 0
				c.errorCount = 0
			}

		}
	}
}
