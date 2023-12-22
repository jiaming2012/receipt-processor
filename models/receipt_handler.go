package models

type ReceiptHandler struct {
	ProjectDir string
	Bindings   []HandlerBinding
}

func NewReceiptHandler(projectDir string) *ReceiptHandler {
	return &ReceiptHandler{
		ProjectDir: projectDir,
	}
}

func (h *ReceiptHandler) Add(path string, processor ReceiptProcessor) {
	h.Bindings = append(h.Bindings, HandlerBinding{
		Path:      path,
		Processor: processor,
	})
}
