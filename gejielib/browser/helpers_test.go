package browser

import (
	"testing"
)

func TestParseCommerceSite(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected SupportedCommerceSite
	}{
		{
			name:     "HappyPath_MercadoLibre",
			url:      "https://www.mercadolibre.com.pe/item",
			expected: MercadoLibre,
		},
		{
			name:     "HappyPath_Dhgate",
			url:      "https://www.dhgate.com/wholesale/search.do?act=&dspm=pcen.sp.searclick.1.Ru0wR5AfTn5h6srAxwNR%26resource_id%3D&sus=&searchkey=cat+toys",
			expected: DHGate,
		},
		{
			name:     "ErrorPath_InvalidURL",
			url:      "not a valid url",
			expected: NotSupported,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			site := ParseCommerceSite(tt.url)
			if site != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, site)
			}
		})
	}
}
