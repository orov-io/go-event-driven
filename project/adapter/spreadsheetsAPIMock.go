package adapter

import (
	"context"
	"sync"
)

type SpreadsheetsAPIMock struct {
	mock         sync.Mutex
	AppendedRows map[string][][]string
}

func NewSpreadsheetsAPIMock() *SpreadsheetsAPIMock {
	return &SpreadsheetsAPIMock{
		mock:         sync.Mutex{},
		AppendedRows: map[string][][]string{},
	}
}

func (r *SpreadsheetsAPIMock) AppendRow(ctx context.Context, sheetName string, row []string) error {
	r.mock.Lock()
	defer r.mock.Unlock()

	if len(r.AppendedRows[sheetName]) == 0 {
		r.AppendedRows[sheetName] = [][]string{row}
	} else {
		r.AppendedRows[sheetName] = append(r.AppendedRows[sheetName], row)
	}

	return nil
}
