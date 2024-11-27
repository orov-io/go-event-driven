package tests_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"tickets/adapter"
	"tickets/service"
	"time"

	"github.com/lithammer/shortuuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	printSheet   = "tickets-to-print"
	refoundSheet = "tickets-to-refund"
)

var confirmedTicket = TicketStatus{
	TicketID: shortuuid.New(),
	Status:   "confirmed",
	Price: Money{
		Amount:   "50.00",
		Currency: "USD",
	},
	Email:     "truman@capote.com",
	BookingID: shortuuid.New(),
}

var canceledTicket = TicketStatus{
	TicketID: shortuuid.New(),
	Status:   "canceled",
	Price: Money{
		Amount:   "50.00",
		Currency: "USD",
	},
	Email:     "truman@capote.com",
	BookingID: shortuuid.New(),
}

func TestComponent(t *testing.T) {

	mocks := runService(t)
	waitForHttpServer(t)
	sendTicketsStatus(t, TicketsStatusRequest{
		Tickets: []TicketStatus{
			confirmedTicket,
			canceledTicket,
		},
	})
	assertReceiptForTicketIssued(t, mocks.Receipts, confirmedTicket)
	assertSpreadsheetRowForTicketIssued(t, mocks.Spreadsheets, confirmedTicket)
	assertSpreadsheetRowForTicketCanceled(t, mocks.Spreadsheets, canceledTicket)
}

func runService(t *testing.T) adapter.ClientMocks {
	t.Helper()
	svc, serviceMocks := service.DefaultMock()
	go func() {
		assert.NoError(t, svc.Run())
	}()

	return serviceMocks
}

func waitForHttpServer(t *testing.T) {
	t.Helper()

	require.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			resp, err := http.Get("http://localhost:8080/health")
			if !assert.NoError(t, err) {
				return
			}
			defer resp.Body.Close()

			if assert.Less(t, resp.StatusCode, 300, "API not ready, http status: %d", resp.StatusCode) {
				return
			}
		},
		time.Second*10,
		time.Millisecond*50,
	)
}

func assertReceiptForTicketIssued(t *testing.T, receiptsService *adapter.ReceiptsServiceMock, ticket TicketStatus) {
	assert.EventuallyWithT(
		t,
		func(collectT *assert.CollectT) {
			issuedReceipts := len(receiptsService.IssuedReceipts)
			t.Log("issued receipts", issuedReceipts)

			assert.Greater(collectT, issuedReceipts, 0, "no receipts issued")
		},
		10*time.Second,
		100*time.Millisecond,
	)

	var receipt adapter.IssueReceiptRequest
	var ok bool
	for _, issuedReceipt := range receiptsService.IssuedReceipts {
		if issuedReceipt.TicketID != ticket.TicketID {
			continue
		}
		receipt = issuedReceipt
		ok = true
		break
	}
	require.Truef(t, ok, "receipt for ticket %s not found", ticket.TicketID)

	assert.Equal(t, ticket.TicketID, receipt.TicketID)
	assert.Equal(t, ticket.Price.Amount, receipt.Price.Amount)
	assert.Equal(t, ticket.Price.Currency, receipt.Price.Currency)
}

func assertSpreadsheetRowForTicketIssued(t *testing.T, spreadsheetsAPI *adapter.SpreadsheetsAPIMock, ticket TicketStatus) {
	assert.EventuallyWithT(
		t,
		func(collectT *assert.CollectT) {
			appendedRows := len(spreadsheetsAPI.AppendedRows[printSheet])
			t.Logf("[%s]appended Rows %d", printSheet, appendedRows)

			assert.Greater(collectT, appendedRows, 0, "no appended rows to ", printSheet)
		},
		10*time.Second,
		100*time.Millisecond,
	)

	var row []string
	var ok bool
	for _, appendedRow := range spreadsheetsAPI.AppendedRows[printSheet] {
		if appendedRow[0] != ticket.TicketID {
			continue
		}
		row = appendedRow
		ok = true
		break
	}
	require.Truef(t, ok, "row for ticket %s not found", ticket.TicketID)

	assert.Equal(t, ticket.TicketID, row[0])
	assert.Equal(t, ticket.Email, row[1])
	assert.Equal(t, ticket.Price.Amount, row[2])
	assert.Equal(t, ticket.Price.Currency, row[3])
}

func assertSpreadsheetRowForTicketCanceled(t *testing.T, spreadsheetsAPI *adapter.SpreadsheetsAPIMock, ticket TicketStatus) {
	assert.EventuallyWithT(
		t,
		func(collectT *assert.CollectT) {
			appendedRows := len(spreadsheetsAPI.AppendedRows[refoundSheet])
			t.Logf("[%s]appended Rows %d", refoundSheet, appendedRows)
			assert.Greater(collectT, appendedRows, 0, "no appended rows to ", refoundSheet)
		},
		10*time.Second,
		100*time.Millisecond,
	)

	var row []string
	var ok bool
	for _, appendedRow := range spreadsheetsAPI.AppendedRows[refoundSheet] {
		if appendedRow[0] != ticket.TicketID {
			continue
		}
		row = appendedRow
		ok = true
		break
	}
	require.Truef(t, ok, "row for ticket %s not found", ticket.TicketID)

	assert.Equal(t, ticket.TicketID, row[0])
	assert.Equal(t, ticket.Email, row[1])
	assert.Equal(t, ticket.Price.Amount, row[2])
	assert.Equal(t, ticket.Price.Currency, row[3])
}

type TicketsStatusRequest struct {
	Tickets []TicketStatus `json:"tickets"`
}

type TicketStatus struct {
	TicketID  string `json:"ticket_id"`
	Status    string `json:"status"`
	Price     Money  `json:"price"`
	Email     string `json:"customer_email"`
	BookingID string `json:"booking_id"`
}

type Money struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

func sendTicketsStatus(t *testing.T, req TicketsStatusRequest) {
	t.Helper()

	payload, err := json.Marshal(req)
	require.NoError(t, err)

	correlationID := shortuuid.New()

	ticketIDs := make([]string, 0, len(req.Tickets))
	for _, ticket := range req.Tickets {
		ticketIDs = append(ticketIDs, ticket.TicketID)
	}

	httpReq, err := http.NewRequest(
		http.MethodPost,
		"http://localhost:8080/tickets-status",
		bytes.NewBuffer(payload),
	)
	require.NoError(t, err)

	httpReq.Header.Set("Correlation-ID", correlationID)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
