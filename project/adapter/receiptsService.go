package adapter

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ThreeDotsLabs/go-event-driven/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients/receipts"
)

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, request IssueReceiptRequest) (IssueReceiptResponse, error)
}

type IssueReceiptRequest struct {
	TicketID string `json:"ticket_id"`
	Price    Money  `json:"price"`
}

type Money struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type IssueReceiptResponse struct {
	ReceiptNumber string    `json:"number"`
	IssuedAt      time.Time `json:"issued_at"`
}

type ReceiptsClient struct {
	clients *clients.Clients
}

func NewReceiptsClient(clients *clients.Clients) ReceiptsClient {
	return ReceiptsClient{
		clients: clients,
	}
}

func (c ReceiptsClient) IssueReceipt(ctx context.Context, request IssueReceiptRequest) (IssueReceiptResponse, error) {
	receiptsResp, err := c.clients.Receipts.PutReceiptsWithResponse(ctx, receipts.PutReceiptsJSONRequestBody{
		Price: receipts.Money{
			MoneyAmount:   request.Price.Amount,
			MoneyCurrency: request.Price.Currency,
		},
		TicketId: request.TicketID,
	})
	if err != nil {
		return IssueReceiptResponse{}, err
	}

	switch receiptsResp.StatusCode() {
	case http.StatusOK:
		return IssueReceiptResponse{
			ReceiptNumber: receiptsResp.JSON200.Number,
			IssuedAt:      receiptsResp.JSON200.IssuedAt,
		}, nil

	case http.StatusCreated:
		return IssueReceiptResponse{
			ReceiptNumber: receiptsResp.JSON201.Number,
			IssuedAt:      receiptsResp.JSON201.IssuedAt,
		}, nil

	default:
		return IssueReceiptResponse{}, fmt.Errorf("unexpected status code: %v", receiptsResp.StatusCode())
	}

}
