package adapter

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

const (
	TypeMetadataKey = "type"
)

func NewEventBus(pub message.Publisher) (*cqrs.EventBus, error) {
	return cqrs.NewEventBusWithConfig(
		pub,
		cqrs.EventBusConfig{
			GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
				return params.EventName, nil
			},
			Marshaler: cqrs.JSONMarshaler{
				GenerateName: cqrs.StructName,
			},
			OnPublish: func(params cqrs.OnEventSendParams) error {
				params.Message.Metadata.Set(TypeMetadataKey, params.EventName)
				return nil
			},
		},
	)
}

type TicketBookingCanceled struct {
	TicketID      string       `json:"ticket_id"`
	CustomerEmail string       `json:"customer_email"`
	Price         MoneyPayload `json:"price"`
}

type TicketBookingConfirmed struct {
	TicketID      string       `json:"ticket_id"`
	CustomerEmail string       `json:"customer_email"`
	Price         MoneyPayload `json:"price"`
}

type MoneyPayload struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}
