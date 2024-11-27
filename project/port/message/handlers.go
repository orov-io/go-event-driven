package message

import (
	"cmp"
	"context"
	"tickets/adapter"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

func (mrr *MessageRouterRunner) issueReceiptHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler(
		"issueReceiptHandler",
		func(ctx context.Context, event *adapter.TicketBookingConfirmed) error {
			_, err := mrr.clients.Receipts.IssueReceipt(ctx, adapter.IssueReceiptRequest{
				TicketID: event.TicketID,
				Price: adapter.Money{
					Amount:   event.Price.Amount,
					Currency: event.Price.Currency,
				},
			})
			return err
		},
	)
}

func (mrr *MessageRouterRunner) printTicketHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler(
		"printTicketHandler",
		func(ctx context.Context, event *adapter.TicketBookingConfirmed) error {
			return mrr.clients.Spreadsheets.AppendRow(
				ctx,
				"tickets-to-print",
				[]string{
					event.TicketID,
					event.CustomerEmail,
					event.Price.Amount,
					cmp.Or(event.Price.Currency, "USD"),
				},
			)
		},
	)
}

func (mrr *MessageRouterRunner) refundTicketHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler(
		"refundTicketHandler",
		func(ctx context.Context, event *adapter.TicketBookingCanceled) error {
			return mrr.clients.Spreadsheets.AppendRow(
				ctx,
				"tickets-to-refund",
				[]string{
					event.TicketID,
					event.CustomerEmail,
					event.Price.Amount,
					cmp.Or(event.Price.Currency, "USD"),
				},
			)
		},
	)
}
