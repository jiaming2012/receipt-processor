package custom

import "regexp"

var DateTimeParse map[string]string
var ReceiptRegex map[string]*regexp.Regexp

func init() {
	ReceiptRegex = make(map[string]*regexp.Regexp)
	DateTimeParse = make(map[string]string)
	ReceiptRegex["Restaurant Depot"] = regexp.MustCompile(`\d{2}-\d{2}-\d{2} \d{2}:\d{2}`)
	ReceiptRegex["Total Units"] = regexp.MustCompile(`TOTAL[\t\s]+UNITS[\t\s]+ENTERED[\t\s]+(\d+)`)
	ReceiptRegex["Total Cases"] = regexp.MustCompile(`TOTAL[\t\s]+CASES[\t\s]+ENTERED[\t\s]+(\d+)`)
	ReceiptRegex["Total Purchases"] = regexp.MustCompile(`TOTAL[\t\s]+ITEMS[\t\s]+RUNG[\t\s]+UP[ \t](\d+)`)
	ReceiptRegex["Additional Purchases"] = regexp.MustCompile(`^ITEMS[\t\s]+RUNG[\t\s]+UP[\t\s]+(\d+)`)
	ReceiptRegex["Subtotal"] = regexp.MustCompile(`SUBTOTAL[ \t]+\$?[0-9]+(\.[0-9][0-9])?`)
	ReceiptRegex["Purchase Delimiter"] = regexp.MustCompile(`(?:(?:CASES)\s+(\d+)\s+)?(?:UNITS)\s+(-?\d+)`)
	DateTimeParse["Restaurant Depot"] = "01-02-06 15:04"
}
