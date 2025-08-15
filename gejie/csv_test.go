package gejie

import (
	"encoding/csv"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestCreateMeliProductCsv(t *testing.T) {
	// Create test data
	testProducts := []MeliProduct{
		{
			Title: "Test Product 1",
			Price: Price{
				AmountCents:    99990,
				AmountCurrency: "S/",
			},
			Url:                "https://example.com/product1",
			ReviewCount:        uint32Ptr(42),
			Rating:             float32Ptr(4.5),
			ImageUrls:          []string{"https://example.com/img1.jpg", "https://example.com/img2.jpg"},
			SoldMoreThan:       uint32Ptr(100),
			DescriptionContent: "This is a test product description",
			Images:             []string{"https://example.com/img3.jpg"},
			StoreInfo: MeliStoreInfo{
				Name:                 "Test Store",
				Url:                  "https://example.com/store",
				LogoImageSrc:         "https://example.com/logo.jpg",
				LogoImageSrcOriginal: "https://example.com/logo_original.jpg",
			},
		},
		{
			Title: "Test Product 2",
			Price: Price{
				AmountCents:    5000,
				AmountCurrency: "S/",
			},
			Url:                "https://example.com/product2",
			ReviewCount:        nil,
			Rating:             nil,
			ImageUrls:          []string{},
			SoldMoreThan:       nil,
			DescriptionContent: "Another test product",
			Images:             []string{},
			StoreInfo: MeliStoreInfo{
				Name:                 "Another Store",
				Url:                  "https://example.com/store2",
				LogoImageSrc:         "",
				LogoImageSrcOriginal: "",
			},
		},
	}

	// Test with valid data
	t.Run("Create CSV with valid products", func(t *testing.T) {
		searchUrl := "https://listado.mercadolibre.com.pe/unit-test"
		searchUrlParsed, _ := url.Parse(searchUrl)
		if len(searchUrlParsed.Path) > 0 && searchUrlParsed.Path[0] == '/' {
			searchUrlParsed.Path = searchUrlParsed.Path[1:]
		}
		fmt.Printf("searchUrl path: %s\n", searchUrlParsed.Path)

		err := CreateMeliProductCsv(testProducts, searchUrlParsed.Path)
		if err != nil {
			t.Fatalf("CreateMeliProductCsv failed: %v", err)
		}

		// Find the created file
		files, err := filepath.Glob("csv_files/" + searchUrlParsed.Path + "-*.csv")
		if err != nil {
			t.Fatalf("Failed to glob for CSV files: %v", err)
		}
		if len(files) == 0 {
			t.Fatal("No CSV file was created")
		}

		// Use the first (and should be only) file
		csvFile := files[0]
		// defer os.Remove(csvFile) // Clean up

		// Verify file contents
		file, err := os.Open(csvFile)
		if err != nil {
			t.Fatalf("Failed to open created CSV file: %v", err)
		}
		defer file.Close()

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil {
			t.Fatalf("Failed to read CSV file: %v", err)
		}

		// Check header
		expectedHeader := []string{
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

		if len(records) < 2 {
			t.Fatal("CSV should have at least header and one data row")
		}

		if !slicesEqual(records[0], expectedHeader) {
			t.Errorf("Header mismatch. Expected: %v, Got: %v", expectedHeader, records[0])
		}

		// Check first product data
		expectedRow1 := []string{
			"Test Product 1",
			"99990",
			"S/",
			"https://example.com/product1",
			"42",
			"4.50",
			"https://example.com/img1.jpg;https://example.com/img2.jpg",
			"100",
			"This is a test product description",
			"https://example.com/img3.jpg",
			"Test Store",
			"https://example.com/store",
			"https://example.com/logo.jpg",
			"https://example.com/logo_original.jpg",
		}

		if !slicesEqual(records[1], expectedRow1) {
			t.Errorf("First row mismatch. Expected: %v, Got: %v", expectedRow1, records[1])
		}

		// Check second product data (with nil values)
		expectedRow2 := []string{
			"Test Product 2",
			"5000",
			"S/",
			"https://example.com/product2",
			"",
			"",
			"",
			"",
			"Another test product",
			"",
			"Another Store",
			"https://example.com/store2",
			"",
			"",
		}

		if !slicesEqual(records[2], expectedRow2) {
			t.Errorf("Second row mismatch. Expected: %v, Got: %v", expectedRow2, records[2])
		}

		// Verify filename contains epoch timestamp
		if !strings.Contains(csvFile, searchUrlParsed.Path+"-") {
			t.Errorf("Filename should contain base filename: %s", csvFile)
		}
		if !strings.HasSuffix(csvFile, ".csv") {
			t.Errorf("Filename should end with .csv: %s", csvFile)
		}
	})

	// Test with empty product list
	t.Run("Create CSV with empty products", func(t *testing.T) {
		baseFilename := "empty_test"
		err := CreateMeliProductCsv([]MeliProduct{}, baseFilename)
		if err != nil {
			t.Fatalf("CreateMeliProductCsv failed with empty products: %v", err)
		}

		// Find the created file
		files, err := filepath.Glob("csv_files/" + baseFilename + "-*.csv")
		if err != nil {
			t.Fatalf("Failed to glob for CSV files: %v", err)
		}
		if len(files) == 0 {
			t.Fatal("No CSV file was created for empty products")
		}

		csvFile := files[0]
		defer os.Remove(csvFile)

		// Verify file contains only header
		file, err := os.Open(csvFile)
		if err != nil {
			t.Fatalf("Failed to open created CSV file: %v", err)
		}
		defer file.Close()

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil {
			t.Fatalf("Failed to read CSV file: %v", err)
		}

		if len(records) != 1 {
			t.Errorf("Empty CSV should have only header row, got %d rows", len(records))
		}
	})

	// Test with special characters in data
	t.Run("Create CSV with special characters", func(t *testing.T) {
		specialProducts := []MeliProduct{
			{
				Title: "Product with \"quotes\" and, commas",
				Price: Price{
					AmountCents:    1000,
					AmountCurrency: "S/",
				},
				Url:                "https://example.com/product",
				ReviewCount:        nil,
				Rating:             nil,
				ImageUrls:          []string{},
				SoldMoreThan:       nil,
				DescriptionContent: "Description with\nnewlines and\ttabs",
				Images:             []string{},
				StoreInfo: MeliStoreInfo{
					Name:                 "Store & Co.",
					Url:                  "https://example.com/store",
					LogoImageSrc:         "",
					LogoImageSrcOriginal: "",
				},
			},
		}

		baseFilename := "special_chars_test"
		err := CreateMeliProductCsv(specialProducts, baseFilename)
		if err != nil {
			t.Fatalf("CreateMeliProductCsv failed with special characters: %v", err)
		}

		// Clean up
		files, err := filepath.Glob("csv_files/" + baseFilename + "-*.csv")
		if err == nil && len(files) > 0 {
			os.Remove(files[0])
		}
	})
}

// Helper function to compare slices
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// Helper function to create uint32 pointer
func uint32Ptr(v uint32) *uint32 {
	return &v
}

// Helper function to create float32 pointer
func float32Ptr(v float32) *float32 {
	return &v
}

// Test filename generation with epoch timestamp
func TestCreateMeliProductCsvFilename(t *testing.T) {
	baseFilename := "test"

	// Capture time before function call
	beforeTime := time.Now().Unix()

	err := CreateMeliProductCsv([]MeliProduct{}, baseFilename)
	if err != nil {
		t.Fatalf("CreateMeliProductCsv failed: %v", err)
	}

	// Capture time after function call
	afterTime := time.Now().Unix()

	// Find the created file
	files, err := filepath.Glob("csv_files/" + baseFilename + "-*.csv")
	if err != nil {
		t.Fatalf("Failed to glob for CSV files: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("No CSV file was created")
	}

	csvFile := files[0]
	defer os.Remove(csvFile)

	// Extract timestamp from filename
	filenameParts := strings.Split(csvFile, "-")
	if len(filenameParts) < 2 {
		t.Fatalf("Filename format incorrect: %s", csvFile)
	}

	timestampStr := strings.TrimSuffix(filenameParts[len(filenameParts)-1], ".csv")
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		t.Fatalf("Failed to parse timestamp from filename: %v", err)
	}

	// Verify timestamp is within expected range
	if timestamp < beforeTime || timestamp > afterTime {
		t.Errorf("Timestamp %d should be between %d and %d", timestamp, beforeTime, afterTime)
	}
}
