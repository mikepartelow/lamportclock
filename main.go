package main

import (
	"fmt"
	"log/slog"
	"os"
	"sort"

	"github.com/rodaine/table"
)

type Message struct {
	AbsoluteId int // Absolute Id of the message. Messages with lower AbsoluteIds were sent before higher ones.
	Clock      int // Sender's Clock
	SenderId   int // Sender's Node Id
}

type Event struct {
	Clock      int
	Message    Message
	ReceiverId int
}

type EventLogger struct {
	EventLog []Event
	Logger   *slog.Logger
}

func (l *EventLogger) Log(e Event) {
	l.Logger.Debug("received event")
	l.EventLog = append(l.EventLog, e)
}

type Node struct {
	Clock       int
	EventLogger *EventLogger
	Id          int
	Lamport     bool
	Logger      *slog.Logger
}

func (n *Node) Receive(m Message) {
	if n.Lamport {
		n.Clock = max(n.Clock, m.Clock)
	}
	n.Clock++

	n.EventLogger.Log(Event{
		Clock:   n.Clock,
		Message: m,
	})
}

func (n *Node) Send(dest *Node, absoluteId int) {
	logger := n.Logger.With("node", n.Id, "absolute id", absoluteId)

	n.Clock++

	m := Message{
		AbsoluteId: absoluteId,
		Clock:      n.Clock,
		SenderId:   n.Id,
	}
	dest.Receive(m)
	logger.Debug("sent")
}

func InitLogger() *slog.Logger {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
	return logger
}

func Run(lamport bool) {
	logger := InitLogger()

	eventLogger := EventLogger{Logger: logger}

	var nodes []*Node

	for i := 0; i < 3; i++ {
		n := &Node{
			EventLogger: &eventLogger,
			Id:          i,
			Lamport:     lamport,
			Logger:      logger,
		}
		nodes = append(nodes, n)
	}

	// node0 sends msg to node2 with clock skewed into the future
	nodes[0].Clock = 10
	nodes[0].Send(nodes[2], 1)

	// node2 sends msg to node1
	nodes[2].Send(nodes[1], 2)

	// node1 sends msg to node0.
	nodes[1].Send(nodes[0], 3)

	// we created a graph of causality: node0 causes an event on node2, which causes an event on node1, which causes an event on node0
	// the order of the events is important, because the content of the messages could depend on state changed by receipt of previous message

	sort.Slice(eventLogger.EventLog, func(i, j int) bool {
		return eventLogger.EventLog[i].Message.AbsoluteId < eventLogger.EventLog[j].Message.AbsoluteId
	})

	tbl := table.New("AbsoluteId", "SenderId", "producer.Clock", "event.Clock")

	for _, e := range eventLogger.EventLog {
		tbl.AddRow(e.Message.AbsoluteId, e.Message.SenderId, e.Message.Clock, e.Clock)
	}

	tbl.Print()
	fmt.Println("Lamport:", lamport)
	fmt.Println("---")
}

func main() {
	Run(false)
	Run(true)
}
