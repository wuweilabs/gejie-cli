package gejie

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

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
		"Price_Amount_Cents",
		"Price_Currency",
		"URL",
		"Review_Count",
		"Rating",
		"Image_URLs",
		"Sold_More_Than",
		"Description_Content",
		"Images",
		"Store_Name",
		"Store_URL",
		"Store_Logo_Image_Src",
		"Store_Logo_Image_Src_Original",
	}

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

		// Join image URLs with semicolon separator
		imageUrls := ""
		if len(product.ImageUrls) > 0 {
			imageUrls = strings.Join(product.ImageUrls, ";")
		}

		images := ""
		if len(product.Images) > 0 {
			images = strings.Join(product.Images, ";")
		}

		row := []string{
			product.Title,
			strconv.Itoa(product.Price.AmountCents),
			product.Price.AmountCurrency,
			product.Url,
			reviewCount,
			rating,
			imageUrls,
			soldMoreThan,
			product.DescriptionContent,
			images,
			product.StoreInfo.Name,
			product.StoreInfo.Url,
			product.StoreInfo.LogoImageSrc,
			product.StoreInfo.LogoImageSrcOriginal,
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}
