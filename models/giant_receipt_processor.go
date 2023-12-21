package models

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	service_models "jiaming2012/receipt-processor/services/models"
)

type GiantReceiptProcessor struct{}

func (p *GiantReceiptProcessor) Process(receiptText string) (service_models.ProcessedReceiptData, error) {
	re := regexp.MustCompile(`Of\t(.+)`)
	match := re.FindStringSubmatch(receiptText)
	storeName := match[1]

	re = regexp.MustCompile(`(\d+.+\n.+, VA \d+)`)
	match = re.FindStringSubmatch(receiptText)
	address := match[1]

	re = regexp.MustCompile(`Store\tTelephone: (\(\d+\) \d+-\d+)`)
	match = re.FindStringSubmatch(receiptText)
	telephone := match[1]

	re = regexp.MustCompile(`Store \d+\t(\d+/\d+/\d+)\t(\d+.\d+am)`)
	match = re.FindStringSubmatch(receiptText)
	dateTimeStr := match[1] + " " + match[2]
	dateTime, err := time.Parse("1/2/06 3.04pm", dateTimeStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing date time: %v", err)
	}

	re = regexp.MustCompile(`(.+)\t(\d+.\d+)`)
	processedText := preprocess(receiptText)
	matches := re.FindAllStringSubmatch(processedText, -1)
	items := make([]string, len(matches))
	for i, match := range matches {
		items[i] = match[1] + ": " + match[2]
	}

	re = regexp.MustCompile(`TAX\t\t(\d+.\d+)`)
	match = re.FindStringSubmatch(receiptText)
	taxAmountStr := match[1]
	taxAmount, err := strconv.ParseFloat(strings.Replace(taxAmountStr, " ", "", -1), 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing total amount: %v", err)
	}

	re = regexp.MustCompile(`\*\*\*\* BALANCE\t\t(\d+.\d+)`)
	match = re.FindStringSubmatch(receiptText)
	subtotalAmountStr := match[1]
	totalAmount, err := strconv.ParseFloat(strings.Replace(subtotalAmountStr, " ", "", -1), 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing total amount: %v", err)
	}

	re = regexp.MustCompile(`Payment Type: (.+)`)
	match = re.FindStringSubmatch(receiptText)
	paymentType := match[1]

	re = regexp.MustCompile(`Card: (\*\*\*\*\*\*\*\*\*\*\*\*\d+)`)
	match = re.FindStringSubmatch(receiptText)
	cardDetails := match[1]

	purchaseItems, err := NewGiantPurchaseItem(items)
	if err != nil {
		panic(err)
	}

	return GiantProcessedReceiptData{
		StoreName:   storeName,
		Address:     address,
		Telephone:   telephone,
		DateTime:    dateTime,
		TotalAmount: totalAmount,
		TaxAmount:   taxAmount,
		PaymentType: paymentType,
		CardLast4:   cardDetails,
		Items:       purchaseItems,
	}, nil
}

func preprocess(text string) string {
	// Split the receipt text into lines
	lines := strings.Split(text, "\n")

	// Filter out lines containing "Store"
	filteredLines := []string{}
	for _, line := range lines {
		if strings.Contains(line, "BALANCE") {
			break
		}

		if strings.Contains(line, "TAX") {
			continue
		}

		if !strings.Contains(line, "Store") {
			filteredLines = append(filteredLines, line)
		}
	}

	// Join the filtered lines back into a string
	return strings.Join(filteredLines, "\n")
}

func NewGiantPurchaseItem(payload []string) ([]service_models.ReceiptPurchaseItem, error) {
	// given a slice of items, create a slice of ReceiptPurchaseItem
	// example payload []string: {WHOLE MILK: 2.69, STRG GAL: 3.99, BLACK BEANS: 1.59, BUY SAVINGS: 0.34, YOU PAY: 1.25, BLACK BEANS: 1.59, BUY SAVINGS: 0.34, YOU PAY: 1.25, OLVOIL 62: 21.99}
	var purchaseItems []service_models.ReceiptPurchaseItem

	k := 0
	for _, item := range payload {
		description, priceStr := strings.Split(item, ":")[0], strings.Split(item, ":")[1]
		if strings.Contains(description, "YOU PAY") {
			continue
		}

		price, err := strconv.ParseFloat(strings.Replace(priceStr, " ", "", -1), 64)
		if err != nil {
			return nil, err
		}

		if strings.Contains(description, "BUY SAVINGS") {
			purchaseItems[k-1].Price = purchaseItems[k-1].Price - price
		} else {
			purchaseItems = append(purchaseItems, service_models.ReceiptPurchaseItem{
				Description: description,
				Price:       price,
				Quantity:    1.0,
			})

			k++
		}

	}

	return purchaseItems, nil
}
