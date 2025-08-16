package gejie

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"
)

type CurrencyRate float32

const PeruvianSolUsdRate CurrencyRate = 0.28
const MexicanPesoUsdRate CurrencyRate = 0.053
const ColombianPesoUsdRate CurrencyRate = 0.00025

// createMeliProductCsv creates a CSV file from a slice of MeliProduct
// The filename will have the current epoch time appended to it
func CreateMeliProductCsv(products []MeliProduct, baseFilename string) error {
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

	// Write CSV header
	header := []string{
		"Title",
		"Original Price Amount",
		"USD Price Amount",
		"URL",
		"Review Count",
		"Rating",
		"Minimum Sold",
		"Minimum Revenue",
		"Description Content",
		"Store Name",
		"Store URL",
	}
	headerChinese := []string{
		"标题",
		"原始价格",
		"USD价格",
		"链接",
		"评论数",
		"评分",
		"销量",
		"最低收入",
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
		originalCurrencyAmount := fmt.Sprintf("%s %f", product.Price.AmountCurrency, amountDecimal)
		//  only Peruvian Soles is supported for now
		usdPriceAmount := fmt.Sprintf("%s %f", currencyAbbrevUnitedStatesDollar, amountDecimal*float32(PeruvianSolUsdRate))
		usdMinimumRevenue := fmt.Sprintf("US$ %f", amountDecimal*float32(PeruvianSolUsdRate))

		row := []string{
			product.Title,
			originalCurrencyAmount,
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

	return nil
}
