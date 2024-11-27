package adapter

import (
	"context"
	"net/http"

	"github.com/ThreeDotsLabs/go-event-driven/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
)

type Clients struct {
	Receipts     ReceiptsService
	Spreadsheets SpreadsheetsAPI
}

func NewClients(addr string) Clients {
	clients, err := clients.NewClients(
		addr,
		func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Correlation-ID", log.CorrelationIDFromContext(ctx))
			return nil
		},
	)
	if err != nil {
		panic(err)
	}

	return Clients{
		Receipts:     NewReceiptsClient(clients),
		Spreadsheets: NewSpreadsheetsClient(clients),
	}
}
