package adapter

import (
	"context"
	"sync"
	"time"

	"github.com/lithammer/shortuuid"
)

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
