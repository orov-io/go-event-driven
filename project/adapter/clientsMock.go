package adapter

type ClientMocks struct {
	Receipts     *ReceiptsServiceMock
	Spreadsheets *SpreadsheetsAPIMock
}

func NewClientsMock() ClientMocks {
	return ClientMocks{
		Receipts:     NewReceiptsServiceMock(),
		Spreadsheets: NewSpreadsheetsAPIMock(),
	}
}
