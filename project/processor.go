package main

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
)

type ProcessorType int

const (
	ProcessorTypeIssueReceipt ProcessorType = iota
	ProcessorTypeAppendToTracker
)

type Processor interface {
	MustStart()
}

type IssueReceiptProcessor struct {
	pType          ProcessorType
	subscriber     message.Subscriber
	receiptsClient ReceiptsClient
	messages       <-chan *message.Message
}

func NewIssueReceiptProcessor(subscriber message.Subscriber, receiptsClient ReceiptsClient) Processor {
	return &IssueReceiptProcessor{
		pType:          ProcessorTypeIssueReceipt,
		subscriber:     subscriber,
		receiptsClient: receiptsClient,
	}
}

func (irp *IssueReceiptProcessor) MustStart() {
	var err error
	irp.messages, err = irp.subscriber.Subscribe(context.Background(), "issue-receipt")
	if err != nil {
		panic(err)
	}

	irp.processMessagesInBackground()
}

func (irp *IssueReceiptProcessor) processMessagesInBackground() {
	go func() {
		for msg := range irp.messages {
			err := irp.receiptsClient.IssueReceipt(msg.Context(), string(msg.Payload))
			if err != nil {
				msg.Nack()
			} else {
				msg.Ack()
			}
		}
	}()
}

type AppendToTrackerProcessor struct {
	pType              ProcessorType
	subscriber         message.Subscriber
	spreadsheetsClient SpreadsheetsClient
	spreadsheetName    string
	messages           <-chan *message.Message
}

func NewAppendToTrackerProcessor(
	subscriber message.Subscriber,
	spreadsheetsClient SpreadsheetsClient,
	spreadsheetName string,
) Processor {
	return &AppendToTrackerProcessor{
		pType:              ProcessorTypeAppendToTracker,
		subscriber:         subscriber,
		spreadsheetsClient: spreadsheetsClient,
		spreadsheetName:    spreadsheetName,
	}
}

func (attp *AppendToTrackerProcessor) MustStart() {
	var err error
	attp.messages, err = attp.subscriber.Subscribe(context.Background(), "append-to-tracker")
	if err != nil {
		panic(err)
	}

	attp.processMessagesInBackground()
}

func (attp *AppendToTrackerProcessor) processMessagesInBackground() {
	go func() {
		for msg := range attp.messages {
			err := attp.spreadsheetsClient.AppendRow(msg.Context(), attp.spreadsheetName, []string{string(msg.Payload)})
			if err != nil {
				msg.Nack()
			} else {
				msg.Ack()
			}
		}
	}()
}
