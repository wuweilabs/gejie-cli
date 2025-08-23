package gejie

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type CurrencyRate float32

const PeruvianSolUsdRate CurrencyRate = 0.28
const MexicanPesoUsdRate CurrencyRate = 0.053
const ColombianPesoUsdRate CurrencyRate = 0.00025

// createMeliProductCsv creates a CSV file from a slice of MeliProduct
// The filename will have the current epoch time appended to it
func CreateMeliProductCsv(products []MeliProduct, baseFilename string) error {
	fmt.Printf("creating csv document (%s) with %d products", baseFilename, len(products))
	// Append epoch timestamp to filename
	epoch := time.Now().Unix()
	filename := fmt.Sprintf("csv_files/%s-%d.csv", baseFilename, epoch)

	// Create csv_files directory if it doesn't exist
	if err := os.MkdirAll("csv_files", 0755); err != nil {
		return fmt.Errorf("failed to create csv_files directory: %w", err)
	}

	// Create the CSV file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	// Create CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Determine currency code for header (assuming all products have same currency)
	var currencyCode string
	if len(products) > 0 {
		currencyCode = string(products[0].Price.CurrencyCode)
	}

	// Write CSV header
	header := []string{
		"Title",
		fmt.Sprintf("Local currency amount (%s)", currencyCode),
		"USD amount",
		"URL",
		"Review count",
		"Rating",
		"Minimum sold",
		"Minimum revenue (USD)",
		"Description content",
		"Store name",
		"Store url",
	}
	headerChinese := []string{
		"标题",
		fmt.Sprintf("本地货币价格 (%s)", currencyCode),
		"美元价格",
		"链接",
		"评论数",
		"评分",
		"销量",
		"最低收入美元",
		"描述内容",
		"店铺名",
		"店铺链接",
	}
	fmt.Print("headers in Chinese: ", headerChinese)

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write product data
	for _, product := range products {
		// Convert fields to strings, handling nil pointers
		reviewCount := ""
		if product.ReviewCount != nil {
			reviewCount = strconv.FormatUint(uint64(*product.ReviewCount), 10)
		}

		rating := ""
		if product.Rating != nil {
			rating = strconv.FormatFloat(float64(*product.Rating), 'f', 2, 32)
		}

		soldMoreThan := ""
		if product.SoldMoreThan != nil {
			soldMoreThan = strconv.FormatUint(uint64(*product.SoldMoreThan), 10)
		}

		amountDecimal := float32(product.Price.AmountCents) / 100
		oCurrencyAmount := fmt.Sprintf("%.2f", amountDecimal)
		//  only Peruvian Soles is supported for now
		usdPriceAmount := fmt.Sprintf("%.2f", amountDecimal*float32(PeruvianSolUsdRate))
		usdMinimumRevenue := fmt.Sprintf("%.2f", amountDecimal*float32(PeruvianSolUsdRate))

		row := []string{
			product.Title,
			oCurrencyAmount,
			usdPriceAmount,
			product.Url,
			reviewCount,
			rating,
			soldMoreThan,
			usdMinimumRevenue,
			product.DescriptionContent,
			product.StoreInfo.Name,
			product.StoreInfo.Url,
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	time.Sleep(2 * time.Second)
	err = PrintCsv(filename)
	if err != nil {
		return fmt.Errorf("failed to print csv: %w", err)
	}

	return nil
}

// PrintCsv prints CSV contents in nicely formatted columns
func PrintCsv(filename string) error {
	fmt.Printf("printing csv file: %s", filename)
	// Open the CSV file
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	// Create CSV reader
	reader := csv.NewReader(file)

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		fmt.Println("CSV file is empty")
		return nil
	}

	// Calculate column widths
	headers := records[0]
	columnWidths := make([]int, len(headers))

	// Find maximum width for each column
	for i, header := range headers {
		columnWidths[i] = len(header)
	}

	// Check data rows for column widths
	for _, record := range records[1:] {
		for i, field := range record {
			if i < len(columnWidths) && len(field) > columnWidths[i] {
				columnWidths[i] = len(field)
			}
		}
	}

	// Print separator line
	printSeparator(columnWidths)

	// Print headers
	printRow(headers, columnWidths, true)

	// Print separator line
	printSeparator(columnWidths)

	// Print data rows
	for _, record := range records[1:] {
		printRow(record, columnWidths, false)
	}

	// Print final separator line
	printSeparator(columnWidths)

	return nil
}

// printRow prints a single row with proper column alignment
func printRow(row []string, columnWidths []int, isHeader bool) {
	for i, field := range row {
		if i < len(columnWidths) {
			// Pad the field to match column width
			paddedField := fmt.Sprintf("%-*s", columnWidths[i], field)
			if isHeader {
				fmt.Printf("| %s ", paddedField)
			} else {
				fmt.Printf("| %s ", paddedField)
			}
		}
	}
	fmt.Println("|")
}

// printSeparator prints a separator line between rows
func printSeparator(columnWidths []int) {
	for _, width := range columnWidths {
		fmt.Printf("+%s", strings.Repeat("-", width+2))
	}
	fmt.Println("+")
}
