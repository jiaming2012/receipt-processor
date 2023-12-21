package models

type HandlerBinding struct {
	Path      string
	Processor ReceiptProcessor
}
