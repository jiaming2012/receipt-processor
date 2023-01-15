package custom

import "regexp"

var DateTimeParse map[string]string
var ReceiptRegex map[string]*regexp.Regexp

func init() {
	ReceiptRegex = make(map[string]*regexp.Regexp)
	DateTimeParse = make(map[string]string)
	ReceiptRegex["Restaurant Depot"] = regexp.MustCompile(`\d{2}-\d{2}-\d{2} \d{2}:\d{2}`)
	ReceiptRegex["Total Units"] = regexp.MustCompile(`TOTAL UNITS ENTERED[ \t](\d+)`)
	ReceiptRegex["Total Cases"] = regexp.MustCompile(`TOTAL CASES ENTERED[ \t](\d+)`)
	ReceiptRegex["Total Purchases"] = regexp.MustCompile(`TOTAL ITEMS RUNG UP[ \t](\d+)`)
	ReceiptRegex["Subtotal"] = regexp.MustCompile(`SUBTOTAL[ \t]+\$?[0-9]+(\.[0-9][0-9])?`)
	DateTimeParse["Restaurant Depot"] = "01-02-06 15:04"
}
