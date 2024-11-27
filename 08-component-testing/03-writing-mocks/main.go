package main

import (
	"context"
	"sync"
	"time"

	"github.com/lithammer/shortuuid/v3"
)

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

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, request IssueReceiptRequest) (IssueReceiptResponse, error)
}

type ReceiptsServiceMock struct {
	mock           sync.Mutex
	IssuedReceipts []IssueReceiptRequest
}

func NewReceiptsServiceMock() *ReceiptsServiceMock {
	return &ReceiptsServiceMock{
		mock:           sync.Mutex{},
		IssuedReceipts: []IssueReceiptRequest{},
	}
}

func (r *ReceiptsServiceMock) IssueReceipt(ctx context.Context, request IssueReceiptRequest) (IssueReceiptResponse, error) {
	r.mock.Lock()
	defer r.mock.Unlock()

	r.IssuedReceipts = append(r.IssuedReceipts, request)
	return IssueReceiptResponse{
		ReceiptNumber: shortuuid.New(),
		IssuedAt:      time.Now(),
	}, nil
}
