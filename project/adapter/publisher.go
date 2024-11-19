package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"tickets/domain/ticket"
	"tickets/middleware/asyncMiddleware"
	"time"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	TypeMetadataKey = "type"
)

const (
	TicketBookingConfirmedEventName = "TicketBookingConfirmed"
	TicketBookingCanceledEventName  = "TicketBookingCanceled"
)

type Publisher struct {
	messagePublisher            message.Publisher
	ticketBookingConfirmedTopic string
	ticketBookingCanceledTopic  string
}

type NewPublisherInfo struct {
	RDB                         *redis.Client
	Logger                      watermill.LoggerAdapter
	TicketBookingConfirmedTopic string
	TicketBookingCanceledTopic  string
}

func MustNewPublisher(info NewPublisherInfo) *Publisher {
	publisher, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: info.RDB,
	}, info.Logger)

	if err != nil {
		panic(fmt.Errorf("unable to create publisher: %w", err))
	}

	return &Publisher{
		messagePublisher:            publisher,
		ticketBookingConfirmedTopic: info.TicketBookingConfirmedTopic,
		ticketBookingCanceledTopic:  info.TicketBookingCanceledTopic,
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

func (p *Publisher) PublishTicketBookingConfirmedEvent(ctx context.Context, ticket ticket.Ticket) {
	payload, _ := json.Marshal(TicketBookingConfirmedEvent{
		Header:        NewEventHeader(TicketBookingConfirmedEventName),
		TicketID:      ticket.ID,
		CustomerEmail: ticket.CustomerEmail,
		Price: MoneyPayload{
			Amount:   ticket.Price.Amount,
			Currency: ticket.Price.Currency,
		},
	})

	msg := message.NewMessage(watermill.NewUUID(), payload)
	msg.Metadata.Set(TypeMetadataKey, TicketBookingConfirmedEventName)

	p.PublishWithCtx(ctx, p.ticketBookingConfirmedTopic, msg)
}

func (p *Publisher) PublishTicketBookingCanceledEvent(ctx context.Context, ticket ticket.Ticket) {
	payload, _ := json.Marshal(TicketBookingConfirmedEvent{
		Header:        NewEventHeader(TicketBookingCanceledEventName),
		TicketID:      ticket.ID,
		CustomerEmail: ticket.CustomerEmail,
		Price: MoneyPayload{
			Amount:   ticket.Price.Amount,
			Currency: ticket.Price.Currency,
		},
	})

	msg := message.NewMessage(watermill.NewUUID(), payload)
	msg.Metadata.Set(TypeMetadataKey, TicketBookingCanceledEventName)

	p.PublishWithCtx(ctx, p.ticketBookingCanceledTopic, msg)
}

func (p *Publisher) PublishWithCtx(ctx context.Context, topic string, messages ...*message.Message) error {
	for _, msg := range messages {
		msg.Metadata.Set(asyncMiddleware.CorrelationIDMetadataKey, log.CorrelationIDFromContext(ctx))
	}

	return p.messagePublisher.Publish(topic, messages...)
}
