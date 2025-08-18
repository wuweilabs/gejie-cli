package gejie

import (
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/zshanhui/gejiezhipin/utils"
)

type Price struct {
	AmountCents  int
	CurrencyCode utils.CurrencyCode
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
	StoreInfo          MeliStoreInfo
}

type MeliStoreInfo struct {
	Name                 string
	Url                  string
	LogoImageSrc         string
	LogoImageSrcOriginal string
}

type BrowserManager struct {
	pw      *playwright.Playwright
	browser playwright.Browser
	context playwright.BrowserContext
}

type BrowserOptions struct {
	Headless    bool
	BlockImages bool
	BlockMedia  bool
	BlockFonts  bool
	UserAgent   string
	Timeout     float64
}

const browserHeadlessMode = false

func DefaultBrowserOptions() *BrowserOptions {
	return &BrowserOptions{
		Headless:    browserHeadlessMode,
		BlockImages: true,
		BlockMedia:  true,
		BlockFonts:  true,
		UserAgent:   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		Timeout:     15000,
	}
}

func NewBrowserManager(opts *BrowserOptions) (*BrowserManager, error) {

	if opts == nil {
		opts = DefaultBrowserOptions()
	}

	pw, err := playwright.Run()
	if err != nil {
		return nil, err
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(opts.Headless),
		Timeout:  playwright.Float(opts.Timeout),
	})
	if err != nil {
		pw.Stop()
		return nil, err
	}

	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String(opts.UserAgent),
	})
	if err != nil {
		browser.Close()
		pw.Stop()
		return nil, err
	}

	if opts.BlockImages || opts.BlockMedia || opts.BlockFonts {
		context.Route("**/*", func(route playwright.Route) {
			rt := route.Request().ResourceType()
			switch rt {
			case "image":
				if opts.BlockImages {
					route.Abort()
					return
				}
			case "media":
				if opts.BlockMedia {
					route.Abort()
					return
				}
			case "font":
				if opts.BlockFonts {
					route.Abort()
					return
				}
			}
			route.Continue()
		})
	}

	return &BrowserManager{
		pw:      pw,
		browser: browser,
		context: context,
	}, nil
}

func (bm *BrowserManager) NewPage() (playwright.Page, error) {
	return bm.context.NewPage()
}

func (bm *BrowserManager) ClosePage(page playwright.Page) {
	if page != nil {
		page.Close()
	}
}

func (bm *BrowserManager) Close() {
	if bm.context != nil {
		bm.context.Close()
	}
	if bm.browser != nil {
		bm.browser.Close()
	}
	if bm.pw != nil {
		bm.pw.Stop()
	}
}

func (bm *BrowserManager) GetContext() playwright.BrowserContext {
	return bm.context
}

func (bm *BrowserManager) GetBrowser() playwright.Browser {
	return bm.browser
}

func RunMeliSearch(searchUrl *string, maxItemsInput int8, createCsv bool) []MeliProduct {
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
		Headless: playwright.Bool(browserHeadlessMode),
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
	err = pageIndex.Locator(productLinksSelector).WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateAttached,
	})
	if err != nil {
		log.Fatalf("product links did not appear: %v", err)
		pageIndex.Close()
	}

	fmt.Print("page loaded, proceeding to scrape links")

	productLinks := ScrapeProductLinksWithPagination(pageIndex, int(maxItemsInput))
	fmt.Printf("\ntotal product links scraped: %d\n", len(productLinks))

	scrapeProducts := []MeliProduct{}
	for _, url := range productLinks {
		product := scrapeProductPage(browser, url)
		if product != nil {
			scrapeProducts = append(scrapeProducts, *product)
		} else {
			fmt.Print("product is nil")
		}
	}
	fmt.Printf("total meli products scraped: %d\n\n", len(scrapeProducts))

	searchUrlParsed, _ := url.Parse(*searchUrl)
	// fmt.Printf("searchUrl path: %s\n", searchUrlParsed.Path)
	if len(searchUrlParsed.Path) > 0 && searchUrlParsed.Path[0] == '/' {
		searchUrlParsed.Path = searchUrlParsed.Path[1:]
	}

	if createCsv {
		fmt.Printf("creating csv for %s, number of products: %d\n", searchUrlParsed.Path, len(scrapeProducts))
		CreateMeliProductCsv(scrapeProducts, searchUrlParsed.Path)
	}

	return scrapeProducts
}

func ScrapeSinglePageProductLinks(page playwright.Page) ([]string, error) {
	productLinks, err := page.Locator(productLinksSelector).All()
	if err != nil {
		return []string{}, fmt.Errorf("could not extract product links: %w", err)
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

	return productLinkUrls, nil
}

func ScrapeProductLinksWithPagination(page playwright.Page, maxItems int) []string {
	allProductLinks := []string{}
	currentPage := 1

	for len(allProductLinks) < maxItems {
		fmt.Printf("scraping page %d...\n", currentPage)

		curPageProductLinks, err := ScrapeSinglePageProductLinks(page)
		if err != nil {
			fmt.Printf("error scraping product links: %v", err)
			return []string{}
		}
		fmt.Printf("found %d product links on page %d\n", len(curPageProductLinks), currentPage)

		remainingItems := maxItems - len(allProductLinks)
		if len(curPageProductLinks) <= remainingItems {
			allProductLinks = append(allProductLinks, curPageProductLinks...)
		} else {
			allProductLinks = append(allProductLinks, curPageProductLinks[:remainingItems]...)
		}

		if len(allProductLinks) >= maxItems {
			fmt.Printf("reached max items (%d), stopping pagination\n", maxItems)
			break
		}

		// next page
		nextButton := page.Locator(string(paginationNextButtonSelector))
		nextExists, err := nextButton.Count()
		if err != nil {
			log.Printf("error checking next page button: %v", err)
			break
		}
		if nextExists == 0 {
			fmt.Printf("no more pages available, stopping pagination")
			break
		}

		// click next button
		err = nextButton.Click()
		if err != nil {
			log.Printf("error clicking next page button: %w", err)
			break
		}

		err = page.Locator(string(productLinksSelector)).Last().WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateAttached,
			Timeout: playwright.Float(5000),
		})
		if err != nil {
			log.Printf("error waiting for next page to load: %v", err)
			break
		}

		// Sleep for 1 second between pages
		time.Sleep(time.Duration(1+rand.Intn(3)) * time.Second)
		currentPage++
	}

	fmt.Printf("total products links scraped across %d pages: %d\n", currentPage, len(allProductLinks))
	return allProductLinks
}

func ScrapeProductPageDirect(url string) *MeliProduct {
	bm, err := NewBrowserManager(&BrowserOptions{
		Headless:    false,
		BlockImages: false,
		BlockMedia:  false,
		BlockFonts:  false,
	})
	if err != nil {
		fmt.Printf("could not create browser manager: %v", err)
	}
	return scrapeProductPage(bm.browser, url)
}

func scrapeProductPage(browser playwright.Browser, url string) *MeliProduct {
	var productPage playwright.Page
	defaultTimeout := float64(8000)

	// create new context allowing media/images to load
	context, err := browser.NewContext()
	if err != nil {
		log.Fatalf("could not create context: %v", err)
	}
	defer context.Close()

	productPage, err = context.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}
	defer productPage.Close()

	_, err = productPage.Goto(url, playwright.PageGotoOptions{
		Timeout: &defaultTimeout,
	})
	if err != nil {
		log.Fatalf("could not goto url: %v", err)
	}

	reviewsContainer := productPage.Locator(string(reviewsContainerSelector))
	ratingCount := ""
	ratingScore := ""
	soldCount := uint32(0)

	// Check if reviews container exists without waiting
	reviewsContainerExists, err := reviewsContainer.Count()
	if err != nil {
		log.Printf("error checking reviews container count: %v", err)
		reviewsContainerExists = 0
	}
	if reviewsContainerExists > 0 {
		// Only wait for visibility if the container exists
		err = reviewsContainer.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(defaultTimeout),
		})
		if err != nil {
			log.Printf("reviews container not visible after waiting: %v", err)
			// Continue with empty values if reviews container is not visible
		} else {
			reviewRating := reviewsContainer.Locator(string(reviewsRatingSelector))
			reviewCount := reviewsContainer.Locator(string(reviewsCountSelector))

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
			soldCount = scrapeSoldCount(productPage)
		}
	} else {
		log.Printf("reviews container not found on page, continuing with empty values")
		// Set default values when no reviews container exists
		ratingCount = ""
		ratingScore = ""
		soldCount = 0
	}

	productName := ""
	nameCount, err := productPage.Locator(string(nameSelector)).Count()
	if nameCount == 0 || err != nil {
		log.Print("product name not found, product page not found")
		return nil
	} else {
		productName, err = productPage.Locator(string(nameSelector)).First().TextContent()
		if err != nil {
			log.Print("productName text not founded")
			return nil
		}
	}

	pageUrl := productPage.URL()
	curCode := utils.DomainToCurrencyCode(utils.Domain(pageUrl))

	amount, err := productPage.Locator(string(priceAmountFractionSelector)).First().TextContent()
	if err != nil {
		log.Fatalf("amount whole not founded")
		return nil
	}

	// amount cent is not always available to scrape
	var amountCentsElem playwright.Locator = nil
	centCount, err := productPage.Locator(string(priceAmountCentSelector)).Count()
	if err != nil {
		log.Fatalf("failed to scrape cent count")
		centCount = 0
	}
	if centCount > 0 {
		amountCentsElem = productPage.Locator(string(priceAmountCentSelector)).First()
	} else {
		fmt.Printf("amount cent not found, continuing with 0\n")
	}

	var amountCentsInt = 0
	if amountCentsElem != nil {
		amountCentsInt = parseCents(amountCentsElem)
	}

	amountInt := utils.StandardizeAmountCents(amount, "")
	fmt.Printf("amount cents parsed: %d, amount whole parsed: %d\n", amountCentsInt, amountInt)

	storeInfo := scrapeStoreInfo(productPage)

	images := ScrapeProductImages(productPage, url)
	fmt.Printf("total product images scraped: %d - first image src: %s\n", len(images), images[0])

	product := MeliProduct{
		Title: productName,
		Price: Price{
			AmountCents:  amountInt + amountCentsInt,
			CurrencyCode: curCode,
		},
		// to be filled in later
		Url:                url,
		ReviewCount:        convertStrUint32(ratingCount),
		Rating:             convertStrToFloat32(ratingScore),
		ImageUrls:          images,
		SoldMoreThan:       &soldCount,
		StoreInfo:          storeInfo,
		DescriptionContent: "",
	}

	return &product
}

func ScrapeProductImages(page playwright.Page, url string) []string {
	var productPage playwright.Page
	if page == nil {
		// direct scrape from url
		opts := DefaultBrowserOptions()
		opts.BlockImages = false
		opts.BlockMedia = false
		opts.BlockFonts = false
		bm, err := NewBrowserManager(opts)
		if err != nil {
			log.Fatalf("could not create browser manager: %v", err)
		}
		defer bm.Close()
		productPage, err = bm.NewPage()
		if err != nil {
			log.Fatalf("could not create page: %v", err)
		}
		defer bm.ClosePage(productPage)
		_, err = productPage.Goto(url, playwright.PageGotoOptions{
			Timeout: playwright.Float(opts.Timeout),
		})
		if err != nil {
			log.Fatalf("could not goto url: %v", err)
		}
		defer bm.ClosePage(productPage)
	} else {
		productPage = page
	}

	images, err := productPage.Locator(string(productImagesSelector)).All()
	if err != nil {
		log.Fatalf("could not extract product images")
	}

	imageUrls := []string{}
	for _, im := range images {
		imageSrc, err := im.GetAttribute("src")
		if err != nil {
			log.Fatalf("could not extract product image url")
		}
		imageUrls = append(imageUrls, imageSrc)
	}
	return imageUrls
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
	storeNameElem := page.Locator(string(storeNameSelector)).First()
	storeName, err := storeNameElem.TextContent()
	if err != nil {
		fmt.Printf("failed to scrape store name: %v", err)
	}
	storeUrlElem := page.Locator(string(storeUrlSelector)).First()
	storeUrl, err := storeUrlElem.GetAttribute("href")
	if err != nil {
		fmt.Printf("failed to scrape store url: %v", err)
		storeUrl = ""
	}
	storeUrl = parseUrlBase(storeUrl)
	storeLogoImageCount, err := page.Locator(string(storeLogoImageSelector)).Count()
	if err != nil {
		fmt.Printf("failed to scrape store logo image does not exist: %v", err)
		storeLogoImageCount = 0
	}
	imageSrc := ""
	if storeLogoImageCount > 0 {
		imageSrc, err = page.Locator(string(storeLogoImageSelector)).First().GetAttribute("data-src")
		if err != nil {
			fmt.Printf("failed to scrape store logo image src: %v", err)
			imageSrc = ""
		}
	}
	return MeliStoreInfo{
		Name:                 storeName,
		Url:                  storeUrl,
		LogoImageSrcOriginal: imageSrc,
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
