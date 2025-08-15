package gejie

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/playwright-community/playwright-go"
)

const exampleMercadoLibreXiaoMi15 = "https://listado.mercadolibre.com.pe/xiaomi-15"
const exampleMercadoLibreKeyboard = "https://listado.mercadolibre.com.pe/teclado-mecanico"

type CurrencyCode string
type CurrencyAbbrev string

const (
	currencyCodePeruvianSoles        CurrencyCode   = "PEN"
	currencyAbbrevPeruvianSoles      CurrencyAbbrev = "S/"
	currencyCodeMexicanPeso          CurrencyCode   = " MXN"
	currencyAbbrevMexicanPeso        CurrencyAbbrev = " Mex$"
	currencyCodeUnitedStatesDollar   CurrencyCode   = "USD"
	currencyAbbrevUnitedStatesDollar CurrencyAbbrev = "US$"
	currencyCodeChineseYuan          CurrencyCode   = "CNY"
	currencyAbbrevChineseYuan        CurrencyAbbrev = "CN¥"
)

const productLinksSelector = ".ui-search-main--only-products div.poly-card__content > h3 > a"

// #root-app > div > div.ui-search-main.ui-search-main--without-header.ui-search-main--only-products > section > div:nth-child(5) > ol > li:nth-child(3) > div > div > div > div.poly-card__content > h3 > a

type Price struct {
	AmountCents    int    // 99990
	AmountCurrency string // S/
}

type MeliProduct struct {
	Title              string
	Price              Price
	Url                string
	ReviewCount        *uint32
	Rating             *float32
	ImageUrls          []string
	SoldMoreThan       *uint32
	DescriptionContent string
	Images             []string
	StoreInfo          MeliStoreInfo
}

type MeliStoreInfo struct {
	Name                 string
	Url                  string
	LogoImageSrc         string
	LogoImageSrcOriginal string
}

func RunMeliSearch(searchUrl *string, maxItemsInput int8) {
	if searchUrl == nil {
		defaultUrl := exampleMercadoLibreKeyboard
		searchUrl = &defaultUrl
	}
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	defer browser.Close()

	// Create new browser context and page
	context, err := browser.NewContext()
	if err != nil {
		log.Fatalf("could not create context: %v", err)
	}
	defer context.Close()

	// Speed up navigation: block heavy resources not needed for scraping links
	_ = context.Route("**/*", func(route playwright.Route) {
		rt := route.Request().ResourceType()
		switch rt {
		case "image", "media", "font":
			_ = route.Abort()
		default:
			_ = route.Continue()
		}
	})

	pageIndex, err := context.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}
	defer pageIndex.Close()

	// Navigate and wait only for DOMContentLoaded to avoid long waits for lazy resources
	_, err = pageIndex.Goto(*searchUrl, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		log.Fatalf("failed to navigate: %v", err)
	}

	// Wait just for product links to appear instead of network idle
	pageIndex.Locator(productLinksSelector).WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateAttached,
	})
	if err != nil {
		log.Fatalf("product links did not appear: %v", err)
	}

	fmt.Print("page loaded, proceeding to scrape links")

	productLinks := ScrapeAllProductLinks(pageIndex)
	fmt.Print("\n\n")
	for _, url := range productLinks {
		fmt.Println(url)
	}
	// fmt.Printf("total product links: %d", len(productLinks))

	// urlFrontier := NewURLFrontier()
	// products := []MeliProduct{}

	// TODO
	// 1. scrape product page
	// 2. scrape all product list pages for links
	// 3. package as easy to use cli tool

	// create another page/tab to scrape products
	pageProducts, err := context.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}
	defer pageIndex.Close()

	scrapeProducts := []MeliProduct{}
	maxItems := len(productLinks)
	if maxItemsInput > 0 {
		maxItems = int(maxItemsInput)
	}
	for i, url := range productLinks {
		if i >= maxItems {
			break
		}
		product := ScrapeProductPage(pageProducts, url)
		if product != nil {
			productJson, err := json.MarshalIndent(product, "", "  ")
			if err != nil {
				log.Printf("failed to marshal product: %v", err)
			} else {
				log.Println("\n\nscraped product:\n", string(productJson))
			}
			scrapeProducts = append(scrapeProducts, *product)
		} else {
			fmt.Print("product is nil")
		}
	}

	fmt.Printf("total meli products scraped: %d", len(scrapeProducts))

	searchUrlParsed, _ := url.Parse(*searchUrl)
	fmt.Printf("searchUrl path: %s\n", searchUrlParsed.Path)
	if len(searchUrlParsed.Path) > 0 && searchUrlParsed.Path[0] == '/' {
		searchUrlParsed.Path = searchUrlParsed.Path[1:]
	}

	CreateMeliProductCsv(scrapeProducts, searchUrlParsed.Path)
}

func ScrapeAllProductLinks(page playwright.Page) []string {
	productLinks, err := page.Locator(productLinksSelector).All()
	if err != nil {
		log.Fatalf("could not extract product links")
	}

	productLinkUrls := []string{}
	baseURL, _ := url.Parse(page.URL())
	for _, productLink := range productLinks {
		linkUrl, err := productLink.GetAttribute("href")

		// skip click1 links since they are not product links
		if strings.HasPrefix(linkUrl, "https://click1") {
			continue
		}
		if err != nil {
			log.Fatalf("could not parse product link url")
			continue
		}
		if linkUrl == "" {
			continue
		}

		ref, err := url.Parse(linkUrl)
		if err != nil {
			// skip malformed URLs
			continue
		}
		abs := baseURL.ResolveReference(ref)
		// strip query params and fragments
		abs.RawQuery = ""
		abs.Fragment = ""

		productLinkUrls = append(productLinkUrls, abs.String())
	}

	return productLinkUrls
}

func ScrapeProductPage(page playwright.Page, url string) *MeliProduct {
	defaultTimeout := float64(5000)
	nameSelector := "h1.ui-pdp-title"
	priceBoxSelector := "#price > div > div.ui-pdp-price__main-container > div.ui-pdp-price__second-line > span > span"
	priceCurrencySelector := ".andes-money-amount__currency-symbol"
	priceAmountFractionSelector := ".andes-money-amount__fraction"
	priceAmountCentSelector := ".andes-money-amount__cents"

	reviewsContainer := page.Locator("div.ui-pdp-header__info > a")

	fmt.Printf("visiting product url: %s", url)
	page.Goto(url, playwright.PageGotoOptions{
		Timeout: playwright.Float(defaultTimeout),
	})

	ratingCount := ""
	ratingScore := ""
	soldCount := uint32(0)
	// Wait for the reviews container to be visible after page navigation
	err := reviewsContainer.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(defaultTimeout),
	})
	if err != nil {
		log.Printf("reviews container not found or not visible: %v", err)
		// Continue with empty values if reviews container is not available
	} else {
		reviewRating := reviewsContainer.Locator("span.ui-pdp-review__rating")
		reviewCount := reviewsContainer.Locator("span.ui-pdp-review__amount")

		ratingCount, err = reviewCount.First().TextContent()
		if err != nil {
			ratingCount = ""
		}
		// Clean the review count by removing parentheses, e.g., "(5)" -> "5"
		ratingCount = cleanReviewCount(ratingCount)
		ratingScore, err = reviewRating.First().TextContent()
		if err != nil {
			ratingScore = ""
		}
		soldCount = scrapeSoldCount(page)
	}

	name, err := page.Locator(nameSelector).First().TextContent()
	if err != nil {
		log.Fatalf("name not founded")
		return nil
	}

	priceBox := page.Locator(priceBoxSelector).First()
	currencyAbbrev, err := priceBox.Locator(priceCurrencySelector).First().TextContent()
	if err != nil {
		log.Fatalf("currency not founded")
		return nil
	}
	amount, err := priceBox.Locator(priceAmountFractionSelector).First().TextContent()
	if err != nil {
		log.Fatalf("amount whole not founded")
		return nil
	}
	fmt.Printf("amount whole found: %s %s\n", currencyAbbrev, amount)

	// handle amount cents scrape, not always available
	amountCentsElem := priceBox.Locator(priceAmountCentSelector).First()
	amountCentsInt := parseCents(amountCentsElem)
	amountInt := standardizeAmount(amount, CurrencyAbbrev(currencyAbbrev))
	fmt.Printf("amount cents parsed: %d, amount whole parsed: %d\n", amountCentsInt, amountInt)

	storeInfo := scrapeStoreInfo(page)
	fmt.Printf("store info: %v", storeInfo)

	product := MeliProduct{
		Title: name,
		Price: Price{
			AmountCents:    amountInt + amountCentsInt,
			AmountCurrency: currencyAbbrev,
		},
		// to be filled in later
		Url:                url,
		ReviewCount:        convertStrUint32(ratingCount),
		Rating:             convertStrToFloat32(ratingScore),
		ImageUrls:          []string{},
		SoldMoreThan:       &soldCount,
		StoreInfo:          storeInfo,
		DescriptionContent: "",
	}

	return &product
}

func scrapeSoldCount(page playwright.Page) uint32 {
	texts, err := page.Locator("div.ui-pdp-header__subtitle > span.ui-pdp-subtitle").AllInnerTexts()
	if err != nil {
		log.Fatalf("failed to scrape sold count")
	}
	// fmt.Printf("scrapeSoldCount / sold text elem: %v", texts)
	soldText := texts[0]
	soldCount := parseSoldCount(soldText)
	return soldCount
}

func scrapeStoreInfo(page playwright.Page) MeliStoreInfo {
	storeNameElem := page.Locator("div.ui-seller-data-header__title-container > h2").First()
	storeName, err := storeNameElem.TextContent()
	if err != nil {
		fmt.Printf("failed to scrape store name: %v", err)
	}
	storeUrlElem := page.Locator("div.ui-seller-data-footer__container > a").First()
	storeUrl, err := storeUrlElem.GetAttribute("href")
	if err != nil {
		fmt.Printf("failed to scrape store url: %v", err)
		storeUrl = ""
	}
	storeUrl = parseUrlBase(storeUrl)

	storeLogoImageSrcOriginal, err := page.Locator("div.ui-seller-data-header__logo-container > a > div > div > img").First().GetAttribute("src")
	if err != nil {
		fmt.Printf("failed to scrape store logo image src: %v", err)
		storeLogoImageSrcOriginal = ""
	}
	// https://www.mercadolibre.com.pe/pagina/negociacionesrepresentaciones
	// item_id=MPE715006378
	// category_id=MPE418448
	// seller_id=833259187
	// client=recoview-selleritems
	// recos_listing=true#origin=pdp
	// &component=sellerData
	// &typeSeller=eshop

	return MeliStoreInfo{
		Name:                 storeName,
		Url:                  storeUrl,
		LogoImageSrcOriginal: storeLogoImageSrcOriginal,
		LogoImageSrc:         "",
	}
}

func parseUrlBase(s string) string {
	if s == "" {
		return ""
	}
	parsedUrl, err := url.Parse(s)
	if err != nil {
		log.Printf("failed to parse url: %v", err)
		return s
	}
	if parsedUrl.Scheme == "" {
		return s
	}
	simpleUrl := parsedUrl.Scheme + "://" + parsedUrl.Host + parsedUrl.Path
	return simpleUrl
}

// Converts a currency abbreviation (e.g., "S/") to its currency code (e.g., "PEN")
func CurrencyAbbrevToCode(abbrev CurrencyAbbrev) CurrencyCode {
	switch abbrev {
	case currencyAbbrevPeruvianSoles:
		return currencyCodePeruvianSoles
	case currencyAbbrevMexicanPeso:
		return currencyCodeMexicanPeso
	case currencyAbbrevUnitedStatesDollar:
		return currencyCodeUnitedStatesDollar
	case currencyAbbrevChineseYuan:
		return currencyCodeChineseYuan
	default:
		return ""
	}
}

// Converts a currency code (e.g., "PEN") to its abbreviation (e.g., "S/")
func CurrencyCodeToAbbrev(code CurrencyCode) CurrencyAbbrev {
	switch code {
	case currencyCodePeruvianSoles:
		return currencyAbbrevPeruvianSoles
	case currencyCodeMexicanPeso:
		return currencyAbbrevMexicanPeso
	case currencyCodeUnitedStatesDollar:
		return currencyAbbrevUnitedStatesDollar
	case currencyCodeChineseYuan:
		return currencyAbbrevChineseYuan
	default:
		return ""
	}
}

func standardizeAmount(amount string, currency CurrencyAbbrev) int {
	if currency == currencyAbbrevPeruvianSoles {
		// In Peruvian Soles, amounts use a decimal point as thousands separator, e.g. "1.234" means 1234.
		// Remove all decimal points before parsing to int.
		amount = strings.ReplaceAll(amount, ".", "")
		amountInt, err := strconv.Atoi(amount)
		if err != nil {
			log.Fatalf("failed to parse amount to int: %v", err)
			return 0
		}
		return amountInt * 100
	}
	return 0
}

func parseCents(amountCentsElem playwright.Locator) int {
	amountCentsInt := 0
	if amountCentsElem == nil {
		fmt.Println("amount cents element not found")
		log.Fatalf("amount cents not founded")
		amountCentsInt = 0
	} else {
		amountCentsText, err := amountCentsElem.TextContent()
		if err != nil {
			amountCentsInt = 0
		} else {
			amountCentsInt, err = strconv.Atoi(amountCentsText)
			if err != nil {
				log.Fatalf("failed to parse amountCents to int: %v", err)
				amountCentsInt = 0
			}
		}
	}
	return amountCentsInt
}

func convertStrToFloat32(s string) *float32 {
	if s != "" {
		if rf, err := strconv.ParseFloat(s, 32); err == nil {
			if err != nil {
				log.Printf("failed to parse rating to float32: %v", err)
				return nil
			}
			rf32 := float32(rf)
			return &rf32
		} else {
			log.Printf("failed to parse rating to float32: %v", err)
			return nil
		}
	} else {
		return nil
	}
}

func convertStrUint32(s string) *uint32 {
	if s != "" {
		// Use ParseUint for unsigned integers to handle the full uint32 range
		if rf, err := strconv.ParseUint(s, 10, 32); err == nil {
			i := uint32(rf)
			return &i
		} else {
			log.Printf("failed to parse string '%s' to uint32: %v", s, err)
			return nil
		}
	} else {
		return nil
	}
}

// Helper function to parse sold count from string like "Nuevo  |  +100 vendidos"
func parseSoldCount(s string) uint32 {
	fmt.Printf("sold text: ^%s^", s)
	parts := strings.Split(s, "|")
	if len(parts) > 1 {
		soldPart := strings.TrimSpace(parts[1])
		soldFields := strings.Fields(soldPart)
		if len(soldFields) > 0 {
			soldNumStr := strings.TrimPrefix(soldFields[0], "+")
			if num, err := strconv.ParseUint(soldNumStr, 10, 32); err == nil {
				return uint32(num)
			}
		}
	}
	return 0
}

// Helper function to clean review count by removing parentheses, e.g., "(5)" -> "5"
func cleanReviewCount(s string) string {
	if s == "" {
		return ""
	}
	// Remove opening and closing parentheses
	s = strings.TrimPrefix(s, "(")
	s = strings.TrimSuffix(s, ")")
	return strings.TrimSpace(s)
}
