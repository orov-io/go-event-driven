package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	TicketBookingConfirmedTopic = "TicketBookingConfirmed"
	TicketBookingCanceledTopic  = "TicketBookingCanceled"
)

type Publisher struct {
	messagePublisher message.Publisher
}

func MustNewPublisher(rdb *redis.Client, logger watermill.LoggerAdapter) *Publisher {
	publisher, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: rdb,
	}, logger)

	if err != nil {
		panic(fmt.Errorf("unable to create publisher: %w", err))
	}

	return &Publisher{
		messagePublisher: publisher,
	}
}

type TicketBookingCanceledEvent struct {
	Header        EventHeader  `json:"header"`
	TicketID      string       `json:"ticket_id"`
	CustomerEmail string       `json:"customer_email"`
	Price         MoneyPayload `json:"price"`
}

type TicketBookingConfirmedEvent struct {
	Header        EventHeader  `json:"header"`
	TicketID      string       `json:"ticket_id"`
	CustomerEmail string       `json:"customer_email"`
	Price         MoneyPayload `json:"price"`
}

type EventHeader struct {
	ID          string `json:"id"`
	EventName   string `json:"event_name"`
	PublishedAt string `json:"published_at"`
}

func NewEventHeader(eventName string) EventHeader {
	return EventHeader{
		ID:          uuid.NewString(),
		EventName:   eventName,
		PublishedAt: time.Now().Format(time.RFC3339),
	}
}

type MoneyPayload struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

func (p *Publisher) PublishTicketBookingConfirmedEvent(ticket Ticket) {
	payload, _ := json.Marshal(TicketBookingConfirmedEvent{
		Header:        NewEventHeader("TicketBookingConfirmed"),
		TicketID:      ticket.ID,
		CustomerEmail: ticket.CustomerEmail,
		Price: MoneyPayload{
			Amount:   ticket.Price.Amount,
			Currency: ticket.Price.Currency,
		},
	})

	p.messagePublisher.Publish(TicketBookingConfirmedTopic, message.NewMessage(watermill.NewUUID(), payload))
}

func (p *Publisher) PublishTicketBookingCanceledEvent(ticket Ticket) {
	payload, _ := json.Marshal(TicketBookingConfirmedEvent{
		Header:        NewEventHeader("TicketBookingCanceled"),
		TicketID:      ticket.ID,
		CustomerEmail: ticket.CustomerEmail,
		Price: MoneyPayload{
			Amount:   ticket.Price.Amount,
			Currency: ticket.Price.Currency,
		},
	})

	p.messagePublisher.Publish(TicketBookingCanceledTopic, message.NewMessage(watermill.NewUUID(), payload))
}
