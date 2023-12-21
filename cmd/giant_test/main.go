package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func preprocess(text string) string {
	// Split the receipt text into lines
	lines := strings.Split(text, "\n")

	// Filter out lines containing "Store"
	filteredLines := []string{}
	for _, line := range lines {
		if strings.Contains(line, "TAX") {
			break
		}

		if !strings.Contains(line, "Store") {
			filteredLines = append(filteredLines, line)
		}
	}

	// Join the filtered lines back into a string
	return strings.Join(filteredLines, "\n")
}

func main() {
	path := "/Users/jamal/projects/yumyums/receipt-processor/receipts/unprocessed/giant/veryfi-ocr-extracted-data-VBFAE-01849.txt"

	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	// extract text from file
	bytes, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	receiptText := string(bytes)

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
	dateTime := match[1] + " " + match[2]

	re = regexp.MustCompile(`(.+)\t(\d+.\d+)`)
	processedText := preprocess(receiptText)
	matches := re.FindAllStringSubmatch(processedText, -1)
	items := make([]string, len(matches))
	for i, match := range matches {
		items[i] = match[1] + ": " + match[2]
	}

	re = regexp.MustCompile(`\*\*\*\* BALANCE\t\t(\d+.\d+)`)
	match = re.FindStringSubmatch(receiptText)
	totalAmount := match[1]

	re = regexp.MustCompile(`Payment Type: (.+)`)
	match = re.FindStringSubmatch(receiptText)
	paymentType := match[1]

	re = regexp.MustCompile(`Card: (\*\*\*\*\*\*\*\*\*\*\*\*\d+)`)
	match = re.FindStringSubmatch(receiptText)
	cardDetails := match[1]

	fmt.Println("Store Name:", storeName)
	fmt.Println("Address:", address)
	fmt.Println("Telephone:", telephone)
	fmt.Println("Date and Time:", dateTime)
	fmt.Println("Total Amount:", totalAmount)
	fmt.Println("Payment Type:", paymentType)
	fmt.Println("Card Details:", cardDetails)

	purchaseItems, err := NewGiantPurchaseItem(items)
	if err != nil {
		panic(err)
	}

	fmt.Println("Purchase Items:", purchaseItems)
}

func NewGiantPurchaseItem(payload []string) ([]GiantPurchaseItem, error) {
	// given a slice of items, create a slice of GiantPurchaseItem
	// example payload []string: {WHOLE MILK: 2.69, STRG GAL: 3.99, BLACK BEANS: 1.59, BUY SAVINGS: 0.34, YOU PAY: 1.25, BLACK BEANS: 1.59, BUY SAVINGS: 0.34, YOU PAY: 1.25, OLVOIL 62: 21.99}
	var purchaseItems []GiantPurchaseItem

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
			purchaseItems = append(purchaseItems, GiantPurchaseItem{
				Description: description,
				Price:       price,
				Quantity:    1.0,
			})

			k++
		}

	}

	return purchaseItems, nil
}

type GiantPurchaseItem struct {
	Sku         string
	Description string
	Price       float64
	Quantity    float64
}

func (g GiantPurchaseItem) String() string {
	if g.Sku == "" {
		return fmt.Sprintf("%s=%.2f", g.Description, g.Price)
	}

	return fmt.Sprintf("%s: %s=%.2f", g.Sku, g.Description, g.Price)
}

func (g GiantPurchaseItem) GetSku() string {
	return g.Sku
}

func (g GiantPurchaseItem) GetDescription() string {
	return g.Description
}

type PurchaseItem interface {
	GetSku() string
	GetDescription() string
}
