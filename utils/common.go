package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
)

func PrintProduct(productStruct interface{}) {
	if productStruct == nil {
		fmt.Println("product is nil")
		return
	}
	productJson, err := json.Marshal(productStruct)
	if err != nil {
		fmt.Printf("failed to marshal product: %v\n", err)
		return
	}
	var productMap map[string]interface{}
	if err := json.Unmarshal(productJson, &productMap); err != nil {
		fmt.Printf("failed to unmarshal product: %v\n", err)
		return
	}
	fmt.Print("\n\n")
	for k, v := range productMap {
		fmt.Printf("%s: %v\n", k, v)
	}
	fmt.Print("\n\n")
}

func GetCountryRegionFromUrl(meliUrl string) Country {
	parsedUrl, err := url.Parse(meliUrl)
	if err != nil {
		log.Printf("failed to parse url: %v", err)
		return ""
	}
	// get the domain part such as com.mx, com, or com.pe
	host := parsedUrl.Host
	parts := strings.Split(host, ".")
	if len(parts) >= 2 {
		// For domains like "mercadolibre.com.mx" or "mercadolibre.com.pe"
		// Return last two parts joined by dot, e.g., "com.mx"
		var domainSubfix string
		if parts[len(parts)-1] == "cl" {
			// For domains like "mercadolibre.cl" (Chile), return .cl
			domainSubfix = parts[len(parts)-1]
		} else {
			domainSubfix = parts[len(parts)-2] + "." + parts[len(parts)-1]
		}
		return DomainToCountry(Domain(domainSubfix))
	}
	return ""
}

type Domain string

const (
	ArgentinaDomain  Domain = "com.ar"
	BoliviaDomain    Domain = "com.bo"
	BrazilDomain     Domain = "com.br"
	ChileDomain      Domain = "cl" // for some reason Chile is just .cl
	ColombiaDomain   Domain = "com.co"
	CostaRicaDomain  Domain = "com.cr"
	DominicanaDomain Domain = "com.do"
	EcuadorDomain    Domain = "com.ec"
	GuatemalaDomain  Domain = "com.gt"
	HondurusDomain   Domain = "com.hn"
	MexicoDomain     Domain = "com.mx"
	NicaraguaDomain  Domain = "com.ni"
	PanamaDomain     Domain = "com.pa"
	ParaguayDomain   Domain = "com.py"
	PeruDomain       Domain = "com.pe"
	ElSalvadorDomain Domain = "com.sv"
	UruguayDomain    Domain = "com.uy"
	VenezuelaDomain  Domain = "com.ve"
)

type Country string

const (
	// supported countries for scraping
	Mexico    Country = "Mexico"
	Peru      Country = "Peru"
	Colombia  Country = "Colombia"
	Chile     Country = "Chile"
	Argentina Country = "Argentina"
	Bolivia   Country = "Bolivia"
	Brasil    Country = "Brasil"

	UnitedStates Country = "United States"
	China        Country = "China"
)

func DomainToCurrencyCode(domain Domain) CurrencyCode {
	c := GetCountryRegionFromUrl(string(domain))
	return CountryRegionToCurrencyCode(c)
}

func DomainToCountry(domain Domain) Country {
	switch domain {
	case ArgentinaDomain:
		return Argentina
	case MexicoDomain:
		return Mexico
	case PeruDomain:
		return Peru
	case ColombiaDomain:
		return Colombia
	case ChileDomain:
		return Chile
	case BoliviaDomain:
		return Bolivia
	// Add more mappings as needed
	default:
		return ""
	}
}

type CurrencyCode string
type CurrencyAbbrev string

const (
	CurrencyCodePeruvianSoles        CurrencyCode   = "PEN"
	CurrencyAbbrevPeruvianSoles      CurrencyAbbrev = "S/"
	CurrencyCodeMexicanPeso          CurrencyCode   = " MXN"
	CurrencyAbbrevMexicanPeso        CurrencyAbbrev = "Mex$"
	CurrencyCodeColombianPeso        CurrencyCode   = "COP"
	CurrencyAbbrevColombianPeso      CurrencyAbbrev = "COP$"
	CurrencyCodeUnitedStatesDollar   CurrencyCode   = "USD"
	CurrencyAbbrevUnitedStatesDollar CurrencyAbbrev = "US$"
	CurrencyCodeChineseYuan          CurrencyCode   = "CNY"
	CurrencyAbbrevChineseYuan        CurrencyAbbrev = "CN¥"
)

func CountryRegionToCurrencyCode(country Country) CurrencyCode {
	switch country {
	case Peru:
		return CurrencyCodePeruvianSoles
	case Mexico:
		return CurrencyCodeMexicanPeso
	case Colombia:
		return CurrencyCodeColombianPeso
	default:
		return ""
	}
}

// Converts a currency abbreviation (e.g., "S/") to its currency code (e.g., "PEN")
func CurrencyAbbrevToCode(abbrev CurrencyAbbrev) CurrencyCode {
	switch abbrev {
	case CurrencyAbbrevPeruvianSoles:
		return CurrencyCodePeruvianSoles
	case CurrencyAbbrevMexicanPeso:
		return CurrencyCodeMexicanPeso
	case CurrencyAbbrevUnitedStatesDollar:
		return CurrencyCodeUnitedStatesDollar
	case CurrencyAbbrevChineseYuan:
		return CurrencyCodeChineseYuan
	default:
		return ""
	}
}

// Converts a currency code (e.g., "PEN") to its abbreviation (e.g., "S/")
func CurrencyCodeToAbbrev(code CurrencyCode) CurrencyAbbrev {
	switch code {
	case CurrencyCodePeruvianSoles:
		return CurrencyAbbrevPeruvianSoles
	case CurrencyCodeMexicanPeso:
		return CurrencyAbbrevMexicanPeso
	case CurrencyCodeUnitedStatesDollar:
		return CurrencyAbbrevUnitedStatesDollar
	case CurrencyCodeChineseYuan:
		return CurrencyAbbrevChineseYuan
	default:
		return ""
	}
}

// StandardizeAmount handles string representation differences in currencies and converts the whole to cents int
func StandardizeAmountCents(amountWhole string, curCode CurrencyCode) int {
	var amountInt int
	var err error
	// separators from left side of decimal should be removed, e.g. "4,333", "4.333" -> "4333"
	amountWhole = strings.ReplaceAll(amountWhole, ".", "")
	amountWhole = strings.ReplaceAll(amountWhole, ",", "")
	amountInt, err = strconv.Atoi(amountWhole)
	if err != nil {
		log.Fatalf("failed to parse amount to int: %v", err)
		return 0
	}

	return amountInt * 100
}
