package utils

import (
	"strings"
	"testing"
)

func TestGetCountryRegionFromUrl(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected Country
	}{
		{
			name:     "Argentina URL",
			url:      "https://www.mercadolibre.com.ar/search?q=laptop",
			expected: Argentina,
		},
		{
			name:     "Mexico URL",
			url:      "https://listado.mercadolibre.com.mx/teclado-mecanico",
			expected: Mexico,
		},
		{
			name:     "Peru URL",
			url:      "https://www.mercadolibre.com.pe/teclado-gamer",
			expected: Peru,
		},
		{
			name:     "Colombia URL",
			url:      "https://articulo.mercadolibre.com.co/iphone-15",
			expected: Colombia,
		},
		{
			name:     "Chile URL with .cl domain",
			url:      "https://www.mercadolibre.cl/notebook",
			expected: Chile,
		},
		{
			name:     "Bolivia URL",
			url:      "https://www.mercadolibre.com.bo/smartphone",
			expected: Bolivia,
		},
		{
			name:     "URL with subdomain",
			url:      "https://api.mercadolibre.com.mx/items",
			expected: Mexico,
		},
		{
			name:     "URL with path and query params",
			url:      "https://www.mercadolibre.com.pe/teclado-gamer-redragon?condition=new&sort=price_asc",
			expected: Peru,
		},
		{
			name:     "URL with fragment",
			url:      "https://articulo.mercadolibre.com.ar/notebook-dell#polycard_client=search",
			expected: Argentina,
		},

		{
			name:     "Invalid URL - should return empty string",
			url:      "not-a-valid-url",
			expected: "",
		},
		{
			name:     "URL with unknown domain - should return empty string",
			url:      "https://www.mercadolibre.com.unknown/search",
			expected: "",
		},
		{
			name:     "URL with single level domain - should return empty string",
			url:      "https://mercadolibre.com/search",
			expected: "",
		},
		{
			name:     "Empty URL - should return empty string",
			url:      "",
			expected: "",
		},
		{
			name:     "URL with only scheme - should return empty string",
			url:      "https://",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetCountryRegionFromUrl(tt.url)
			if result != tt.expected {
				t.Errorf("GetCountryRegionFromUrl(%q) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestDomainToCountry(t *testing.T) {
	tests := []struct {
		domain   Domain
		expected Country
	}{
		{ArgentinaDomain, Argentina},
		{MexicoDomain, Mexico},
		{PeruDomain, Peru},
		{ColombiaDomain, Colombia},
		{ChileDomain, Chile},
		{BoliviaDomain, Bolivia},
		{"com.unknown", ""}, // Unknown domain should return empty string
		{"", ""},            // Empty domain should return empty string
	}

	for _, tt := range tests {
		t.Run(string(tt.domain), func(t *testing.T) {
			result := DomainToCountry(tt.domain)
			if result != tt.expected {
				t.Errorf("DomainToCountry(%q) = %v, want %v", tt.domain, result, tt.expected)
			}
		})
	}
}

func TestGetCountryRegionFromUrlEdgeCases(t *testing.T) {
	// Test with very long URLs
	longURL := "https://" + strings.Repeat("a", 1000) + ".mercadolibre.com.mx/search"
	result := GetCountryRegionFromUrl(longURL)
	if result != Mexico {
		t.Errorf("GetCountryRegionFromUrl with long URL failed, got %v, want %v", result, Mexico)
	}

	// Test with URLs containing special characters
	specialURL := "https://www.mercadolibre.com.pe/search?q=laptop%20gaming&price=1000-2000"
	result = GetCountryRegionFromUrl(specialURL)
	if result != Peru {
		t.Errorf("GetCountryRegionFromUrl with special characters failed, got %v, want %v", result, Peru)
	}

	// Test with URLs containing emojis
	emojiURL := "https://www.mercadolibre.com.ar/search?q=laptop🔥"
	result = GetCountryRegionFromUrl(emojiURL)
	if result != Argentina {
		t.Errorf("GetCountryRegionFromUrl with emoji failed, got %v, want %v", result, Argentina)
	}
}

func BenchmarkGetCountryRegionFromUrl(b *testing.B) {
	testURL := "https://www.mercadolibre.com.mx/search?q=laptop"
	for i := 0; i < b.N; i++ {
		GetCountryRegionFromUrl(testURL)
	}
}

func TestStandardizeAmountCents(t *testing.T) {
	tests := []struct {
		amount   string
		curCode  CurrencyCode
		expected int
	}{
		{"1,234", CurrencyCodeMexicanPeso, 123400},
		{"1.234", CurrencyCodePeruvianSoles, 123400},
		{"1,234.56", CurrencyCodePeruvianSoles, 12345600},
		{"0", CurrencyCodePeruvianSoles, 0},
		{"5", CurrencyCodePeruvianSoles, 500},
		{"999,999", CurrencyCodePeruvianSoles, 99999900},
	}

	for _, tt := range tests {
		result := StandardizeAmountCents(tt.amount, tt.curCode)
		if result != tt.expected {
			t.Errorf("StandardizeAmountCents(%q, %q) = %d, want %d", tt.amount, tt.curCode, result, tt.expected)
		}
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
			abbrev:   CurrencyAbbrevPeruvianSoles,
			expected: CurrencyCodePeruvianSoles,
		},
		{
			name:     "Mexican Peso abbreviation",
			abbrev:   CurrencyAbbrevMexicanPeso,
			expected: CurrencyCodeMexicanPeso,
		},
		{
			name:     "US Dollar abbreviation",
			abbrev:   CurrencyAbbrevUnitedStatesDollar,
			expected: CurrencyCodeUnitedStatesDollar,
		},
		{
			name:     "Chinese Yuan abbreviation",
			abbrev:   CurrencyAbbrevChineseYuan,
			expected: CurrencyCodeChineseYuan,
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
			code:     CurrencyCodePeruvianSoles,
			expected: CurrencyAbbrevPeruvianSoles,
		},
		{
			name:     "Mexican Peso code",
			code:     CurrencyCodeMexicanPeso,
			expected: CurrencyAbbrevMexicanPeso,
		},
		{
			name:     "US Dollar code",
			code:     CurrencyCodeUnitedStatesDollar,
			expected: CurrencyAbbrevUnitedStatesDollar,
		},
		{
			name:     "Chinese Yuan code",
			code:     CurrencyCodeChineseYuan,
			expected: CurrencyAbbrevChineseYuan,
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

func TestStandardizeAmount(t *testing.T) {
	tests := []struct {
		name     string
		amount   string
		curCode  CurrencyCode
		expected int
	}{
		{
			name:     "PEN currency with no decimal points",
			amount:   "1234",
			curCode:  CurrencyCodePeruvianSoles,
			expected: 123400, // 1234 * 100
		},
		{
			name:     "PEN currency with single decimal point as thousands separator",
			amount:   "1.234",
			curCode:  CurrencyCodePeruvianSoles,
			expected: 123400, // removes "." -> 1234 -> 1234 * 100
		},
		{
			name:     "PEN currency with multiple decimal points",
			amount:   "1.234.567",
			curCode:  CurrencyCodePeruvianSoles,
			expected: 123456700, // removes all "." -> 1234567 -> 1234567 * 100
		},
		{
			name:     "PEN currency with zero amount",
			amount:   "0",
			curCode:  CurrencyCodePeruvianSoles,
			expected: 0,
		},
		{
			name:     "PEN currency with single digit",
			amount:   "5",
			curCode:  CurrencyCodePeruvianSoles,
			expected: 500,
		},
		{
			name:     "return 123400 for USD currency",
			amount:   "1234",
			curCode:  "USD",
			expected: 123400,
		},
		{
			name:     "return 123400 for MXN currency",
			amount:   "1.234",
			curCode:  " MXN",
			expected: 123400,
		},
		{
			name:     "return 123400 for empty currency",
			amount:   "1234",
			curCode:  "",
			expected: 123400,
		},
		{
			name:     "PEN currency with large amount",
			amount:   "1.234.567.890",
			curCode:  CurrencyCodePeruvianSoles,
			expected: 123456789000, // removes all "." -> 1234567890 -> 1234567890 * 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StandardizeAmountCents(tt.amount, tt.curCode)
			if result != tt.expected {
				t.Errorf("standardizeAmount(%q, %q) = %d, expected %d",
					tt.amount, tt.curCode, result, tt.expected)
			}
		})
	}
}
