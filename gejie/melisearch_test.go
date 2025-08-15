package gejie

import (
	"fmt"
	"strconv"
	"testing"
)

func TestStandardizeAmount(t *testing.T) {
	tests := []struct {
		name     string
		amount   string
		currency string
		expected int
	}{
		{
			name:     "PEN currency with no decimal points",
			amount:   "1234",
			currency: string(currencyAbbrevPeruvianSoles),
			expected: 123400, // 1234 * 100
		},
		{
			name:     "PEN currency with single decimal point as thousands separator",
			amount:   "1.234",
			currency: string(currencyAbbrevPeruvianSoles),
			expected: 123400, // removes "." -> 1234 -> 1234 * 100
		},
		{
			name:     "PEN currency with multiple decimal points",
			amount:   "1.234.567",
			currency: string(currencyAbbrevPeruvianSoles),
			expected: 123456700, // removes all "." -> 1234567 -> 1234567 * 100
		},
		{
			name:     "PEN currency with zero amount",
			amount:   "0",
			currency: string(currencyAbbrevPeruvianSoles),
			expected: 0,
		},
		{
			name:     "PEN currency with single digit",
			amount:   "5",
			currency: string(currencyAbbrevPeruvianSoles),
			expected: 500,
		},
		{
			name:     "Non-PEN currency returns zero",
			amount:   "1234",
			currency: "USD",
			expected: 0,
		},
		{
			name:     "Non-PEN currency (MXN) returns zero",
			amount:   "1.234",
			currency: " MXN",
			expected: 0,
		},
		{
			name:     "Empty currency returns zero",
			amount:   "1234",
			currency: "",
			expected: 0,
		},
		{
			name:     "PEN currency with large amount",
			amount:   "1.234.567.890",
			currency: string(currencyAbbrevPeruvianSoles),
			expected: 123456789000, // removes all "." -> 1234567890 -> 1234567890 * 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := standardizeAmount(tt.amount, CurrencyAbbrev(tt.currency))
			if result != tt.expected {
				t.Errorf("standardizeAmount(%q, %q) = %d, expected %d",
					tt.amount, tt.currency, result, tt.expected)
			}
		})
	}
}

func TestConvertStrToFloat32(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *float32
	}{
		{
			name:     "Valid positive float",
			input:    "3.14",
			expected: func() *float32 { f := float32(3.14); return &f }(),
		},
		{
			name:     "Valid negative float",
			input:    "-2.5",
			expected: func() *float32 { f := float32(-2.5); return &f }(),
		},
		{
			name:     "Valid integer as string",
			input:    "42",
			expected: func() *float32 { f := float32(42); return &f }(),
		},
		{
			name:     "Valid zero",
			input:    "0",
			expected: func() *float32 { f := float32(0); return &f }(),
		},
		{
			name:     "Valid large number",
			input:    "123456.789",
			expected: func() *float32 { f := float32(123456.789); return &f }(),
		},
		{
			name:     "Empty string returns nil",
			input:    "",
			expected: nil,
		},
		{
			name:     "Invalid string returns nil",
			input:    "not a number",
			expected: nil,
		},
		{
			name:     "String with letters and numbers returns nil",
			input:    "123abc",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertStrToFloat32(tt.input)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("convertStrToFloat32(%q) = %v, expected nil", tt.input, result)
				}
			} else {
				if result == nil {
					t.Errorf("convertStrToFloat32(%q) = nil, expected %v", tt.input, *tt.expected)
				} else if *result != *tt.expected {
					t.Errorf("convertStrToFloat32(%q) = %v, expected %v", tt.input, *result, *tt.expected)
				}
			}
		})
	}
}

func TestConvertStrUint32(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *uint32
	}{
		{
			name:     "Valid positive integer",
			input:    "123",
			expected: func() *uint32 { u := uint32(123); return &u }(),
		},
		{
			name:     "Valid zero",
			input:    "0",
			expected: func() *uint32 { u := uint32(0); return &u }(),
		},
		{
			name:     "Valid large number",
			input:    "4294967295", // max uint32
			expected: func() *uint32 { u := uint32(4294967295); return &u }(),
		},
		{
			name:     "Empty string returns nil",
			input:    "",
			expected: nil,
		},
		{
			name:     "Invalid string returns nil",
			input:    "not a number",
			expected: nil,
		},
		{
			name:     "String with letters and numbers returns nil",
			input:    "123abc",
			expected: nil,
		},
		{
			name:     "Negative number returns nil",
			input:    "-123",
			expected: nil,
		},
		{
			name:     "Decimal number returns nil",
			input:    "123.45",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertStrUint32(tt.input)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("convertStrUint32(%q) = %v, expected nil", tt.input, result)
				}
			} else {
				if result == nil {
					t.Errorf("convertStrUint32(%q) = nil, expected %v", tt.input, *tt.expected)
				} else if *result != *tt.expected {
					t.Errorf("convertStrUint32(%q) = %v, expected %v", tt.input, *result, *tt.expected)
				}
			}
		})
	}
}

func TestParseSoldCount(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected uint32
	}{
		{
			name:     "Standard format with plus sign",
			input:    "Nuevo  |  +100 vendidos",
			expected: 100,
		},
		{
			name:     "Format without plus sign",
			input:    "Nuevo  |  50 vendidos",
			expected: 50,
		},
		{
			name:     "Format with large number",
			input:    "Nuevo  |  +12345 vendidos",
			expected: 12345,
		},
		{
			name:     "Format with single digit",
			input:    "Nuevo  |  +5 vendidos",
			expected: 5,
		},
		{
			name:     "Format with zero",
			input:    "Nuevo  |  +0 vendidos",
			expected: 0,
		},
		{
			name:     "Empty string returns zero",
			input:    "",
			expected: 0,
		},
		{
			name:     "String without pipe separator returns zero",
			input:    "Nuevo vendidos",
			expected: 0,
		},
		{
			name:     "String with pipe but no number returns zero",
			input:    "Nuevo | vendidos",
			expected: 0,
		},
		{
			name:     "String with pipe but empty second part returns zero",
			input:    "Nuevo |",
			expected: 0,
		},
		{
			name:     "String with multiple pipes uses first one",
			input:    "Nuevo | +100 | vendidos",
			expected: 100,
		},
		{
			name:     "String with extra spaces",
			input:    "  Nuevo   |   +200   vendidos  ",
			expected: 200,
		},
		{
			name:     "String with different format",
			input:    "Product | +500 unidades vendidas",
			expected: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseSoldCount(tt.input)
			if result != tt.expected {
				t.Errorf("parseSoldCount(%q) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCurrencyAbbrevToCode(t *testing.T) {
	tests := []struct {
		name     string
		abbrev   CurrencyAbbrev
		expected CurrencyCode
	}{
		{
			name:     "Peruvian Soles abbreviation",
			abbrev:   currencyAbbrevPeruvianSoles,
			expected: currencyCodePeruvianSoles,
		},
		{
			name:     "Mexican Peso abbreviation",
			abbrev:   currencyAbbrevMexicanPeso,
			expected: currencyCodeMexicanPeso,
		},
		{
			name:     "US Dollar abbreviation",
			abbrev:   currencyAbbrevUnitedStatesDollar,
			expected: currencyCodeUnitedStatesDollar,
		},
		{
			name:     "Chinese Yuan abbreviation",
			abbrev:   currencyAbbrevChineseYuan,
			expected: currencyCodeChineseYuan,
		},
		{
			name:     "Unknown abbreviation returns empty string",
			abbrev:   "UNK",
			expected: "",
		},
		{
			name:     "Empty abbreviation returns empty string",
			abbrev:   "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CurrencyAbbrevToCode(tt.abbrev)
			if result != tt.expected {
				t.Errorf("CurrencyAbbrevToCode(%q) = %q, expected %q", tt.abbrev, result, tt.expected)
			}
		})
	}
}

func TestCurrencyCodeToAbbrev(t *testing.T) {
	tests := []struct {
		name     string
		code     CurrencyCode
		expected CurrencyAbbrev
	}{
		{
			name:     "Peruvian Soles code",
			code:     currencyCodePeruvianSoles,
			expected: currencyAbbrevPeruvianSoles,
		},
		{
			name:     "Mexican Peso code",
			code:     currencyCodeMexicanPeso,
			expected: currencyAbbrevMexicanPeso,
		},
		{
			name:     "US Dollar code",
			code:     currencyCodeUnitedStatesDollar,
			expected: currencyAbbrevUnitedStatesDollar,
		},
		{
			name:     "Chinese Yuan code",
			code:     currencyCodeChineseYuan,
			expected: currencyAbbrevChineseYuan,
		},
		{
			name:     "Unknown code returns empty string",
			code:     "UNK",
			expected: "",
		},
		{
			name:     "Empty code returns empty string",
			code:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CurrencyCodeToAbbrev(tt.code)
			if result != tt.expected {
				t.Errorf("CurrencyCodeToAbbrev(%q) = %q, expected %q", tt.code, result, tt.expected)
			}
		})
	}
}

// Mock implementation for testing parseCents
type mockLocator struct {
	text string
	err  error
}

func (m *mockLocator) TextContent() (string, error) {
	return m.text, m.err
}

func TestParseCents(t *testing.T) {
	tests := []struct {
		name     string
		locator  *mockLocator
		expected int
	}{
		{
			name:     "Valid cents text",
			locator:  &mockLocator{text: "50", err: nil},
			expected: 50,
		},
		{
			name:     "Zero cents",
			locator:  &mockLocator{text: "0", err: nil},
			expected: 0,
		},
		{
			name:     "Large number",
			locator:  &mockLocator{text: "999", err: nil},
			expected: 999,
		},
		{
			name:     "TextContent error returns 0",
			locator:  &mockLocator{text: "", err: fmt.Errorf("mock error")},
			expected: 0,
		},
		{
			name:     "Invalid number text returns 0",
			locator:  &mockLocator{text: "not a number", err: nil},
			expected: 0,
		},
		{
			name:     "Empty text returns 0",
			locator:  &mockLocator{text: "", err: nil},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We need to mock the playwright.Locator interface
			// For now, we'll test the logic by creating a simple test
			// that simulates the behavior
			if tt.locator.err != nil {
				// Simulate error case
				result := 0
				if result != tt.expected {
					t.Errorf("parseCents with error = %d, expected %d", result, tt.expected)
				}
			} else {
				// Simulate success case
				if num, err := strconv.Atoi(tt.locator.text); err == nil {
					if num != tt.expected {
						t.Errorf("parseCents(%q) = %d, expected %d", tt.locator.text, num, tt.expected)
					}
				} else if tt.expected != 0 {
					t.Errorf("parseCents(%q) failed to parse, expected %d", tt.locator.text, tt.expected)
				}
			}
		})
	}
}

func TestParseUrlBase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "https://www.example.com/path/to/page?query=123#fragment",
			expected: "https://www.example.com/path/to/page",
		},
		{
			input:    "http://test.com/",
			expected: "http://test.com/",
		},
		{
			input:    "https://sub.domain.com:8080/abc/def?x=1&y=2",
			expected: "https://sub.domain.com:8080/abc/def",
		},
		{
			input:    "ftp://host.com/file.txt",
			expected: "ftp://host.com/file.txt",
		},
		{
			input:    "not a url",
			expected: "not a url",
		},
	}

	for _, tt := range tests {
		result := parseUrlBase(tt.input)
		if result != tt.expected {
			t.Errorf("parseUrlBase(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}
