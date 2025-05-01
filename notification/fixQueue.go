package notification

import (
	"strings"
)

type queueCommand struct {
	cmd    string
	value  string
	result chan string // used for returning result from String()
}

type FixedQueue struct {
	commands chan queueCommand
}

func NewFixedQueue(limit int) *FixedQueue {
	q := &FixedQueue{
		commands: make(chan queueCommand),
	}
	go q.run(limit)
	return q
}

func (q *FixedQueue) run(limit int) {
	items := make([]string, 0, limit)

	for cmd := range q.commands {
		switch cmd.cmd {
		case "add":
			if len(items) == limit {
				items = items[1:]
			}
			items = append(items, cmd.value)
		case "string":
			cmd.result <- strings.Join(items, "\n")
		case "clear":
			items = make([]string, 0, limit)
		case "stringAndClear":
			resp := strings.Join(items, "\n")
			items = make([]string, 0, limit)
			cmd.result <- resp
		}
	}
}

// Public API

func (q *FixedQueue) Add(item string) {
	q.commands <- queueCommand{cmd: "add", value: item}
}

func (q *FixedQueue) String() string {
	resp := make(chan string)
	q.commands <- queueCommand{cmd: "string", result: resp}
	return <-resp
}

func (q *FixedQueue) Clear() {
	q.commands <- queueCommand{cmd: "clear"}
}

func (q *FixedQueue) StringAndClear() string {
	resp := make(chan string)
	q.commands <- queueCommand{cmd: "stringAndClear", result: resp}
	return <-resp
}
