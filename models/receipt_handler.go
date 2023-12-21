package models

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"jiaming2012/receipt-processor/services"
	service_models "jiaming2012/receipt-processor/services/models"
	"jiaming2012/receipt-processor/utils"
)

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

func (h *ReceiptHandler) Run(db *gorm.DB) {
	for _, binding := range h.Bindings {
		log.Debugf("Running processor for path: %s", binding.Path)
		unprocessedFiles := filepath.Join(h.ProjectDir, "receipts", "unprocessed", binding.Path)
		moveToDir := filepath.Join(h.ProjectDir, "receipts", "processed", binding.Path)

		err := processDirectory(unprocessedFiles, moveToDir, db, binding.Processor.Process)
		if err != nil {
			log.Errorf("error walking directory: %v", err)
			continue
		}
	}
}

// processDirectory walks the directory and calls the callback function for each file.
// The callback function is passed the contents of the file.
// The file is moved to the processedDir after the callback function is called.
func processDirectory(dir string, processedDir string, db *gorm.DB, callback func(string) (service_models.ProcessedReceiptData, error)) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if err != nil {
			return fmt.Errorf("error walking directory: %v", err)
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("error opening file: %v", err)
		}

		defer file.Close()

		bytes, err := io.ReadAll(file)
		if err != nil {
			return fmt.Errorf("error reading file: %v", err)
		}

		if !info.IsDir() {
			receiptData, callbackErr := callback(string(bytes))
			if callbackErr != nil {
				return fmt.Errorf("error in callback: %v", callbackErr)
			}

			// Validate the receipt data
			if err := services.ValidateReceiptData(receiptData); err != nil {
				return fmt.Errorf("error validating receipt data: %v", err)
			}

			// Transform the receipt data

			// Save data to database
			if err := services.SaveReceiptData(receiptData, db); err != nil {
				return fmt.Errorf("error saving receipt data: %v", err)
			}

			err := utils.MoveFileToDirectory(path, processedDir)
			if err != nil {
				return fmt.Errorf("error moving file: %v", err)
			}
		}

		return nil
	})
}
