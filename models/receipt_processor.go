package models

type ReceiptProcessor interface {
	Process(receiptText string) (ProcessedReceiptData, error)
}
