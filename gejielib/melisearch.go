package gejie

import (
	"log"
	"log/slog"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/zshanhui/gejiezhipin/gejielib/adapters"
	br "github.com/zshanhui/gejiezhipin/gejielib/browser"
	"github.com/zshanhui/gejiezhipin/gejielib/meli"
	"github.com/zshanhui/gejiezhipin/gejielib/uf"
	"github.com/zshanhui/gejiezhipin/utils"
)

func RunMeliSearch(searchUrl *string, opts br.CmdOptions) []meli.MeliProduct {
	if searchUrl == nil {
		defaultUrl := exampleMercadoLibreKeyboard
		searchUrl = &defaultUrl
	}
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	defer pw.Stop()

	browserOpts := br.DefaultBrowserOptions()
	browserOpts.Headless = false
	if opts.HeadlessMode {
		browserOpts.Headless = true
	}

	browser, err := br.CreateBrowser(pw, browserOpts, opts)
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

	slog.Info("page loaded, proceeding to scrape links")

	productLinks := ScrapeProductLinksWithPaginationHelper(pageIndex, int(opts.MaxItems))
	if len(productLinks) == 0 {
		slog.Error("no product links found, potential blocked by mercadolibre due to local vpn use, stopping")
		return nil
	}
	slog.Info("total product links scraped", "count", len(productLinks))

	numWorkers := opts.Workers
	if numWorkers <= 0 {
		numWorkers = 1
	}
	// jobs := make(chan string)
	results := make(chan meli.MeliProduct, len(productLinks))

	urlFrontier := uf.NewURLFrontier()
	urlFrontier.BulkAdd(productLinks)

	var wg sync.WaitGroup
	workerUrlFrontier := func() {
		defer wg.Done()
		for {
			ur, ok := urlFrontier.GetNext()
			if !ok {
				// no more pending urls to scrape
				return
			}
			urlFrontier.MarkVisited(ur)
			// prd := meli.ScrapeProductPageDummy(browser, ur)
			prd := scrapeProductPage(browser, ur)
			if prd != nil {
				results <- *prd
			} else {
				urlFrontier.MarkFailed(ur)
			}
		}
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go workerUrlFrontier()
	}

	go func() { wg.Wait(); close(results) }()

	scrapeProducts := []meli.MeliProduct{}
	for prd := range results {
		scrapeProducts = append(scrapeProducts, prd)
	}
	slog.Info("total meli products scraped", "count", len(scrapeProducts))
	if len(scrapeProducts) == 0 {
		slog.Error("no products scraped successfully; all product pages failed")
		return nil
	}
	slog.Info("first product scraped")
	utils.PrintProduct(scrapeProducts[0])

	searchUrlParsed, _ := url.Parse(*searchUrl)
	// fmt.Printf("searchUrl path: %s\n", searchUrlParsed.Path)
	if len(searchUrlParsed.Path) > 0 && searchUrlParsed.Path[0] == '/' {
		searchUrlParsed.Path = searchUrlParsed.Path[1:]
	}

	if opts.CreateCsv {
		slog.Info("creating csv", "path", searchUrlParsed.Path, "count", len(scrapeProducts))
		CreateMeliProductCsv(scrapeProducts, searchUrlParsed.Path, false)
	}

	return scrapeProducts
}

func ScrapeProductLinksWithPaginationHelper(page playwright.Page, maxItems int) []string {
	opts := adapters.ScrapeProductLinkPageOpts{
		MaxItems:               maxItems,
		ProductLinkSelector:    productLinksSelector,
		PaginationNextSelector: string(paginationNextButtonSelector),
	}
	return adapters.ScrapeProductLinksWithPagination(page, opts)
}

func ScrapeProductPageDirect(url string, opts br.CmdOptions) *meli.MeliProduct {
	bm, err := br.NewBrowserManager(&br.BrowserOptions{
		Headless:    opts.HeadlessMode,
		BlockImages: false,
		BlockMedia:  false,
		BlockFonts:  false,
	}, opts)
	if err != nil {
		slog.Error("could not create browser manager", "error", err)
	}

	return scrapeProductPage(bm.GetBrowser(), url)
}

func scrapeProductPage(browser playwright.Browser, url string) *meli.MeliProduct {
	gejieConfig := br.DefaultGejieConfig()
	productTimer := utils.NewTimer("Individual Product Page")
	defer productTimer.LogElapsed()

	var productPage playwright.Page
	defaultTimeout := float64(gejieConfig.BrowserTimeout)

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
	var soldCount *uint32

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
		soldCount = nil
	}

	productName := ""
	nameCount, err := productPage.Locator(string(nameSelector)).Count()
	if nameCount == 0 || err != nil {
		slog.Error("product name not found, product page not found")
		return nil
	} else {
		productName, err = productPage.Locator(string(nameSelector)).First().TextContent()
		if err != nil {
			slog.Error("productName text not founded")
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
		slog.Info("amount cent not found, continuing with 0")
	}

	var amountCentsInt = 0
	if amountCentsElem != nil {
		amountCentsInt = parseCents(amountCentsElem)
	}

	amountInt := utils.StandardizeAmountCents(amount, "")
	slog.Info("parsed amounts", "amountCents", amountCentsInt, "amountWhole", amountInt)

	storeInfo := scrapeStoreInfo(productPage)

	images := ScrapeProductImages(productPage, url)
	slog.Info("total product images scraped", "count", len(images), "firstImage", images[0])

	content := scrapeProductDescription(productPage)

	product := meli.MeliProduct{
		Title: productName,
		Price: meli.Price{
			AmountCents:  amountInt + amountCentsInt,
			CurrencyCode: curCode,
		},
		// to be filled in later
		Url:         url,
		ReviewCount: convertStrUint32(ratingCount),
		Rating:      convertStrToFloat32(ratingScore),
		ImageUrls:   images,
		// SoldMoreThan is only an estimate and can be inaccurate
		SoldMoreThan:       soldCount,
		EstimatedSoldCount: estimateSoldCount(convertStrUint32(ratingCount)),
		StoreInfo:          storeInfo,
		DescriptionContent: content,
	}

	// to add some randomness so it seems more like natural browsing
	randMs := 1000 + rand.Intn(2200) // random ms between 1000 and 3199
	slog.Info("delaying for milliseconds", "ms", randMs)
	time.Sleep(time.Duration(randMs) * time.Millisecond)

	return &product
}

func estimateSoldCount(reviewCount *uint32) *uint32 {
	// percentage of average buyers that review across all product categories
	// some product have higher review percents like electronics or books
	reviewAverage := 0.04
	if reviewCount == nil {
		return nil
	}
	est := float64(*reviewCount) / reviewAverage
	result := uint32(est)
	return &result
}

func ScrapeProductImages(page playwright.Page, url string) []string {
	var productPage playwright.Page
	if page == nil {
		// direct scrape from url
		opts := br.DefaultBrowserOptions()
		opts.BlockImages = false
		opts.BlockMedia = false
		opts.BlockFonts = false
		bm, err := br.NewBrowserManager(opts, br.CmdOptions{
			Browser: br.BrowserTypeFirefox,
		})
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

func scrapeProductDescription(page playwright.Page) string {
	locator := page.Locator(string(productDescriptionSelector))
	count, err := locator.Count()
	if err != nil || count == 0 {
		// If the element does not exist or error occurs, return empty string
		return ""
	}
	content, err := locator.InnerHTML()
	if err != nil {
		slog.Error("failed to scrape product description", "error", err)
		return ""
	}
	return content
}

func scrapeSoldCount(page playwright.Page) *uint32 {
	texts, err := page.Locator(string(minimumSoldCountSelector)).AllInnerTexts()
	if err != nil {
		log.Fatalf("failed to scrape sold count")
	}
	// fmt.Printf("scrapeSoldCount / sold text elem: %v", texts)
	soldText := texts[0]
	soldCount := parseSoldCount(soldText)
	return soldCount
}

func scrapeStoreInfo(page playwright.Page) meli.MeliStoreInfo {
	storeNameElem := page.Locator(string(storeNameSelector)).First()
	storeName, err := storeNameElem.TextContent()
	if err != nil {
		slog.Error("failed to scrape store name", "error", err)
	}
	storeUrlElem := page.Locator(string(storeUrlSelector)).First()
	storeUrl, err := storeUrlElem.GetAttribute("href")
	if err != nil {
		slog.Error("failed to scrape store url", "error", err)
		storeUrl = ""
	}
	storeUrl = parseUrlBase(storeUrl)
	storeLogoImageCount, err := page.Locator(string(storeLogoImageSelector)).Count()
	if err != nil {
		slog.Error("failed to scrape store logo image does not exist", "error", err)
		storeLogoImageCount = 0
	}
	imageSrc := ""
	if storeLogoImageCount > 0 {
		imageSrc, err = page.Locator(string(storeLogoImageSelector)).First().GetAttribute("data-src")
		if err != nil {
			slog.Error("failed to scrape store logo image src", "error", err)
			imageSrc = ""
		}
	}
	return meli.MeliStoreInfo{
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
		slog.Error("amount cents element not found")
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
func parseSoldCount(s string) *uint32 {
	parts := strings.Split(s, "|")
	if len(parts) > 1 {
		soldPart := strings.TrimSpace(parts[1])
		soldFields := strings.Fields(soldPart)
		// Check if "mil" is present in soldPart and print the result if so
		if strings.Contains(soldPart, "mil") {
			// the sold count is not accurate when it contains "mil"
			slog.Info("soldPart contains 'mil'", "soldPart", soldPart)
			return nil
		}
		if len(soldFields) > 0 {
			soldNumStr := strings.TrimPrefix(soldFields[0], "+")
			if num, err := strconv.ParseUint(soldNumStr, 10, 32); err == nil {
				count := uint32(num)
				return &count
			}
		}
	}
	return nil
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
