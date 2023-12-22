package models

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type GiantReceiptProcessor struct{}

func (p *GiantReceiptProcessor) Process(receiptText string) (ProcessedReceiptData, error) {
	re := regexp.MustCompile(`Giant`)
	match := re.FindStringSubmatch(receiptText)
	storeName := match[0]

	re = regexp.MustCompile(`(\d+.+\n.+, [A-Z]{2} \d+)`)
	match = re.FindStringSubmatch(receiptText)
	address := match[1]

	re = regexp.MustCompile(`Store\sTelephone: (\((\d{3})\) \d{3}-\d{4})`)
	match = re.FindStringSubmatch(receiptText)
	telephone := match[1]

	re = regexp.MustCompile(`Store \d+\t(\d+/\d+/\d+)\t(\d+.\d+am)`)
	match = re.FindStringSubmatch(receiptText)
	dateTimeStr := match[1] + " " + match[2]
	dateTimeStr = strings.Replace(dateTimeStr, ".", ":", -1)
	dateTime, err := time.Parse("1/2/06 3:04pm", dateTimeStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing date time: %v", err)
	}

	re = regexp.MustCompile(`(.+)\t(\d+\.\d+\s\w)(;;.+)?`)
	processedText := preprocess(receiptText)
	matches := re.FindAllStringSubmatch(processedText, -1)
	items := make([]string, len(matches))
	for i, match := range matches {
		items[i] = match[1] + ": " + strings.Join(match[2:], "")
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

	purchaseItems, err := NewGiantPurchaseItems(items)
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
	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Fixes an edge case where @ in the unit price is interpreted as B
		line = strings.Replace(line, " B ", " @ ", -1)
		if strings.Contains(line, "@") {
			// Fixes an edge case where \s is interpreted as \t
			line = strings.ReplaceAll(line, "\t", " ")
		}

		// Append the unit price to the line containing the purchase
		if strings.Contains(line, " @ ") {
			line = fmt.Sprintf("%s;;%s", lines[i+1], line)
			i += 1
		}

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

// func preprocessGiantPurchaseItems(payload []string) []string {
// 	var output []string

// 	for i := 0; i < len(payload) - 1; i++ {
// 		if payload[i+1]
// 	}
// }

func NewGiantPurchaseItems(payload []string) ([]ReceiptPurchaseItem, error) {
	// given a slice of items, create a slice of ReceiptPurchaseItem
	// example payload []string: {WHOLE MILK: 2.69, STRG GAL: 3.99, BLACK BEANS: 1.59, BUY SAVINGS: 0.34, YOU PAY: 1.25, BLACK BEANS: 1.59, BUY SAVINGS: 0.34, YOU PAY: 1.25, OLVOIL 62: 21.99}
	var purchaseItems []ReceiptPurchaseItem

	k := 0
	for _, multiLineItem := range payload {
		items := strings.Split(multiLineItem, ";;")

		var item string
		var unitPrice *string
		if len(items) > 1 {
			item = items[0]
			unitPrice = &items[1]
		} else {
			item = items[0]
		}

		description, priceStr := strings.Split(item, ":")[0], strings.Split(item, ":")[1]
		if strings.Contains(description, "YOU PAY") {
			continue
		}

		re := regexp.MustCompile(`[^0-9.]`)
		priceStr = re.ReplaceAllString(priceStr, "")
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			return nil, fmt.Errorf("NewGiantPurchaseItems: error parsing price: %w", err)
		}

		if strings.Contains(description, "BUY SAVINGS") {
			purchaseItems[k-1].Price = purchaseItems[k-1].Price - price
		} else {
			purchaseItems = append(purchaseItems, ReceiptPurchaseItem{
				Description: description,
				Quantity:    1.0,
				Price:       price,
				UnitPrice:   unitPrice,
			})

			k++
		}

	}

	return purchaseItems, nil
}
