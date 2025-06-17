package transfers

type TransferRequest struct {
	FromAccount string  `json:"source_account_id"`
	ToAccount   string  `json:"destination_account_id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
}

type WebhookEvent struct {
	ID     string `json:"transfer_id"`
	Status string `json:"status"`
}
