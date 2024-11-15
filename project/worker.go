package main

import "context"

type Task int

const (
	TaskIssueReceipt Task = iota
	TaskAppendToTracker
)

type Message struct {
	Task     Task
	TicketID string
}

type Worker struct {
	queue              chan Message
	receiptsClient     ReceiptsClient
	spreadsheetsClient SpreadsheetsClient
}

func NewWorker(receiptsClient ReceiptsClient, spreadsheetsClient SpreadsheetsClient) *Worker {
	return &Worker{
		queue:              make(chan Message, 100),
		receiptsClient:     receiptsClient,
		spreadsheetsClient: spreadsheetsClient,
	}
}

func (w *Worker) Send(msg ...Message) {
	for _, m := range msg {
		w.queue <- m
	}
}

func (w *Worker) Run() {
	for msg := range w.queue {
		switch msg.Task {
		case TaskIssueReceipt:
			w.processTaskIssuerReceipt(msg)
		case TaskAppendToTracker:
			w.processTaskAppendToTracker(msg)
		}
	}
}

func (w *Worker) processTaskIssuerReceipt(msg Message) {
	err := w.receiptsClient.IssueReceipt(context.Background(), msg.TicketID)
	if err != nil {
		w.Send(msg)
	}
}

func (w *Worker) processTaskAppendToTracker(msg Message) {
	err := w.spreadsheetsClient.AppendRow(context.Background(), "tickets-to-print", []string{msg.TicketID})
	if err != nil {
		w.Send(msg)
	}
}
