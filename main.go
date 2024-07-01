package main

import (
	"fmt"
	"log/slog"
	"os"
	"sort"
	"sync"

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
	Done     <-chan bool
	Events   chan Event
	EventLog []Event
	Logger   *slog.Logger
	Wg       *sync.WaitGroup
}

func (l *EventLogger) Log(e Event) {
	l.Events <- e
}

func (l *EventLogger) Consume() {
	logger := l.Logger.With("EventLogger", "consumer")
	l.Wg.Add(1)
	go func() {
		defer l.Wg.Done()
		for {
			select {
			case <-l.Done:
				logger.Warn("done")
				return
			case e := <-l.Events:
				logger.Debug("received event")
				l.EventLog = append(l.EventLog, e)
			}
		}
	}()
}

type Node struct {
	Clock       int
	Done        <-chan bool
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

	clock := n.Clock

	n.EventLogger.Log(Event{
		Clock:   clock,
		Message: m,
	})
}

func (n *Node) Produce(dest *Node, absoluteId int) {
	logger := n.Logger.With("node", n.Id, "producer", true, "absolute id", absoluteId)

	n.Clock++
	clock := n.Clock

	m := Message{
		AbsoluteId: absoluteId,
		Clock:      clock,
		SenderId:   n.Id,
	}
	dest.Receive(m)
	logger.Debug("sent", "absoluteId", m.AbsoluteId)
}

func InitLogger() *slog.Logger {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
	return logger
}

func Run(lamport bool) {
	logger := InitLogger()
	done := make(chan bool)
	var wg sync.WaitGroup

	eventLogger := EventLogger{
		Done:   done,
		Events: make(chan Event),
		Logger: logger,
		Wg:     &wg,
	}
	eventLogger.Consume()

	var nodes []*Node

	for i := 0; i < 3; i++ {
		n := &Node{
			Done:        done,
			EventLogger: &eventLogger,
			Id:          i,
			Lamport:     lamport,
			Logger:      logger,
		}
		nodes = append(nodes, n)
	}

	// node0 sends msg to node2 with clock skewed into the future
	nodes[0].Clock = 10
	nodes[0].Produce(nodes[2], 1)

	// node2 sends msg to node1
	nodes[2].Produce(nodes[1], 2)

	// node1 sends msg to node0.
	nodes[1].Produce(nodes[0], 3)

	// we created a graph of causality: node0 causes an event on node2, which causes an event on node1, which causes an event on node0
	// the order of the events is important, because the content of the messages could depend on state changed by receipt of previous message

	close(done)
	wg.Wait()

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
